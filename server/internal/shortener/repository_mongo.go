package shortener

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const urlsCollection = "urls"

// urlDoc es la representación de persistencia. _id es un ObjectID real; el repo
// traduce a/desde el hex string del domain.ShortURL.
type urlDoc struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	ShortCode     string             `bson:"shortCode"`
	OriginalURL   string             `bson:"originalUrl"`
	OwnerID       string             `bson:"ownerId"`
	IsCustomAlias bool               `bson:"isCustomAlias"`
	IsActive      bool               `bson:"isActive"`
	CreatedAt     time.Time          `bson:"createdAt"`
	UpdatedAt     time.Time          `bson:"updatedAt"`
	TotalClicks   int64              `bson:"totalClicks"`
}

func (d urlDoc) toDomain() *ShortURL {
	return &ShortURL{
		ID:            d.ID.Hex(),
		ShortCode:     d.ShortCode,
		OriginalURL:   d.OriginalURL,
		OwnerID:       d.OwnerID,
		IsCustomAlias: d.IsCustomAlias,
		IsActive:      d.IsActive,
		CreatedAt:     d.CreatedAt,
		UpdatedAt:     d.UpdatedAt,
		TotalClicks:   d.TotalClicks,
	}
}

type mongoURLRepository struct {
	coll *mongo.Collection
}

// NewURLRepository construye el repo y garantiza los índices (único en shortCode,
// secundario en ownerId). Falla en boot si no se pueden crear.
func NewURLRepository(db *mongo.Database) (URLRepository, error) {
	coll := db.Collection(urlsCollection)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "shortCode", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("uniq_shortCode"),
		},
		{
			Keys:    bson.D{{Key: "ownerId", Value: 1}},
			Options: options.Index().SetName("idx_ownerId"),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("shortener: no se pudieron crear los índices: %w", err)
	}

	slog.Info("repositorio de URLs listo", "collection", urlsCollection, "indexes", "uniq_shortCode, idx_ownerId")
	return &mongoURLRepository{coll: coll}, nil
}

func (r *mongoURLRepository) Create(ctx context.Context, url *ShortURL) error {
	doc := urlDoc{
		ShortCode:     url.ShortCode,
		OriginalURL:   url.OriginalURL,
		OwnerID:       url.OwnerID,
		IsCustomAlias: url.IsCustomAlias,
		IsActive:      url.IsActive,
		CreatedAt:     url.CreatedAt,
		UpdatedAt:     url.UpdatedAt,
		TotalClicks:   url.TotalClicks, // 0 al crear; se incrementa con $inc en cada clic
	}

	res, err := r.coll.InsertOne(ctx, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrAliasTaken
		}
		return fmt.Errorf("shortener: error insertando URL: %w", err)
	}
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		url.ID = oid.Hex()
	}
	return nil
}

func (r *mongoURLRepository) FindByCode(ctx context.Context, code string) (*ShortURL, error) {
	var doc urlDoc
	err := r.coll.FindOne(ctx, bson.M{"shortCode": code}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrURLNotFound
		}
		return nil, fmt.Errorf("shortener: error buscando URL por código: %w", err)
	}
	return doc.toDomain(), nil
}

func (r *mongoURLRepository) FindByID(ctx context.Context, id string) (*ShortURL, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrURLNotFound
	}
	var doc urlDoc
	err = r.coll.FindOne(ctx, bson.M{"_id": oid}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrURLNotFound
		}
		return nil, fmt.Errorf("shortener: error buscando URL por id: %w", err)
	}
	return doc.toDomain(), nil
}

// FindByOwner lista las URLs del usuario aplicando, en una sola query de Mongo, el filtro
// de estado (isActive) y la búsqueda parcial sobre shortCode/originalUrl. DECISIÓN: se usa
// $regex case-insensitive (no un text index de Mongo): para el volumen esperado del proyecto
// basta con el índice idx_ownerId existente, que acota primero por dueño; un text index no
// soporta "contains" arbitrario (solo palabras) y sería sobre-ingeniería aquí. El término
// del usuario se escapa con regexp.QuoteMeta para tratarlo como texto literal (anti-inyección).
func (r *mongoURLRepository) FindByOwner(ctx context.Context, ownerID string, listFilter ListFilter, page, limit int) ([]*ShortURL, int64, error) {
	filter := bson.M{"ownerId": ownerID}
	if listFilter.IsActive != nil {
		filter["isActive"] = *listFilter.IsActive
	}
	if search := strings.TrimSpace(listFilter.Search); search != "" {
		rx := primitive.Regex{Pattern: regexp.QuoteMeta(search), Options: "i"}
		filter["$or"] = bson.A{
			bson.M{"shortCode": rx},
			bson.M{"originalUrl": rx},
		}
	}

	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("shortener: error contando URLs del usuario: %w", err)
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))

	cur, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("shortener: error listando URLs del usuario: %w", err)
	}
	defer cur.Close(ctx)

	var docs []urlDoc
	if err := cur.All(ctx, &docs); err != nil {
		return nil, 0, fmt.Errorf("shortener: error decodificando URLs: %w", err)
	}

	urls := make([]*ShortURL, len(docs))
	for i, d := range docs {
		urls[i] = d.toDomain()
	}
	return urls, total, nil
}

// CountActiveByOwner cuenta las URLs activas del usuario (isActive: true). Se
// apoya en el índice idx_ownerId. Las desactivadas no entran en el cupo del plan.
func (r *mongoURLRepository) CountActiveByOwner(ctx context.Context, ownerID string) (int64, error) {
	count, err := r.coll.CountDocuments(ctx, bson.M{"ownerId": ownerID, "isActive": true})
	if err != nil {
		return 0, fmt.Errorf("shortener: error contando URLs activas del usuario: %w", err)
	}
	return count, nil
}

func (r *mongoURLRepository) ExistsByCode(ctx context.Context, code string) (bool, error) {
	count, err := r.coll.CountDocuments(ctx, bson.M{"shortCode": code}, options.Count().SetLimit(1))
	if err != nil {
		return false, fmt.Errorf("shortener: error verificando existencia de código: %w", err)
	}
	return count > 0, nil
}

func (r *mongoURLRepository) Update(ctx context.Context, url *ShortURL) error {
	oid, err := primitive.ObjectIDFromHex(url.ID)
	if err != nil {
		return ErrURLNotFound
	}
	res, err := r.coll.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": bson.M{
		"isActive":  url.IsActive,
		"updatedAt": url.UpdatedAt,
	}})
	if err != nil {
		return fmt.Errorf("shortener: error actualizando URL: %w", err)
	}
	if res.MatchedCount == 0 {
		return ErrURLNotFound
	}
	return nil
}

// Delete elimina por id verificando ownership en el mismo filtro. Si no se borra
// nada (no existe o no es del owner) devuelve ErrNotOwner, para no revelar si el
// id existe (decisión de seguridad, ver criterio de aceptación).
func (r *mongoURLRepository) Delete(ctx context.Context, id, ownerID string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrNotOwner
	}
	res, err := r.coll.DeleteOne(ctx, bson.M{"_id": oid, "ownerId": ownerID})
	if err != nil {
		return fmt.Errorf("shortener: error eliminando URL: %w", err)
	}
	if res.DeletedCount == 0 {
		return ErrNotOwner
	}
	return nil
}

// IncrementClicks suma 1 al contador de forma atómica. Filtra por shortCode (que
// tiene índice único). Si no existe, devuelve ErrURLNotFound. El incremento es fire-and-forget: un fallo
func (r *mongoURLRepository) IncrementClicks(ctx context.Context, shortCode string) error {
	res, err := r.coll.UpdateOne(ctx,
		bson.M{"shortCode": shortCode},
		bson.M{"$inc": bson.M{"totalClicks": 1}},
	)
	if err != nil {
		return fmt.Errorf("shortener: error incrementando contador de clics: %w", err)
	}
	if res.MatchedCount == 0 {
		return ErrURLNotFound
	}
	return nil
}

// SumClicksByOwner agrega el TotalClicks de todas las URLs del usuario con un
// pipeline ($match + $group sum). Si el usuario no tiene URLs, devuelve 0.
func (r *mongoURLRepository) SumClicksByOwner(ctx context.Context, ownerID string) (int64, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{{Key: "ownerId", Value: ownerID}}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "total", Value: bson.D{{Key: "$sum", Value: "$totalClicks"}}},
		}}},
	}
	cur, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, fmt.Errorf("shortener: error agregando clics del usuario: %w", err)
	}
	defer cur.Close(ctx)

	var rows []struct {
		Total int64 `bson:"total"`
	}
	if err := cur.All(ctx, &rows); err != nil {
		return 0, fmt.Errorf("shortener: error decodificando suma de clics: %w", err)
	}
	if len(rows) == 0 {
		return 0, nil
	}
	return rows[0].Total, nil
}

// FindShortCodesByOwner devuelve solo los shortCodes del usuario (proyección),
// usado para purgar el cache de redirección al eliminar una cuenta.
func (r *mongoURLRepository) FindShortCodesByOwner(ctx context.Context, ownerID string) ([]string, error) {
	opts := options.Find().SetProjection(bson.M{"shortCode": 1, "_id": 0})
	cur, err := r.coll.Find(ctx, bson.M{"ownerId": ownerID}, opts)
	if err != nil {
		return nil, fmt.Errorf("shortener: error listando shortCodes del usuario: %w", err)
	}
	defer cur.Close(ctx)

	var rows []struct {
		ShortCode string `bson:"shortCode"`
	}
	if err := cur.All(ctx, &rows); err != nil {
		return nil, fmt.Errorf("shortener: error decodificando shortCodes: %w", err)
	}
	codes := make([]string, len(rows))
	for i, row := range rows {
		codes[i] = row.ShortCode
	}
	return codes, nil
}

// DeleteAllByOwner elimina todas las URLs del usuario en una sola operación.
func (r *mongoURLRepository) DeleteAllByOwner(ctx context.Context, ownerID string) error {
	if _, err := r.coll.DeleteMany(ctx, bson.M{"ownerId": ownerID}); err != nil {
		return fmt.Errorf("shortener: error eliminando las URLs del usuario: %w", err)
	}
	return nil
}
