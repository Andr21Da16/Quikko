package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const usersCollection = "users"

// userDoc es la representación de persistencia.
type userDoc struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Email        string             `bson:"email"`
	PasswordHash string             `bson:"passwordHash"`
	Plan         Plan               `bson:"plan"`
	CreatedAt    time.Time          `bson:"createdAt"`
}

func (d userDoc) toDomain() *User {
	plan := d.Plan

	if plan == "" {
		plan = PlanFree
	}
	return &User{
		ID:           d.ID.Hex(),
		Email:        d.Email,
		PasswordHash: d.PasswordHash,
		Plan:         plan,
		CreatedAt:    d.CreatedAt,
	}
}

// mongoUserRepository implementa UserRepository sobre MongoDB.
type mongoUserRepository struct {
	coll *mongo.Collection
}

// NewUserRepository construye el repo y garantiza el índice único en email.
// Falla en boot si el índice no se puede crear (fail-fast).
func NewUserRepository(db *mongo.Database) (UserRepository, error) {
	coll := db.Collection(usersCollection)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("uniq_email"),
	})
	if err != nil {
		return nil, fmt.Errorf("auth: no se pudo crear el índice único de email: %w", err)
	}

	slog.Info("repositorio de usuarios listo", "collection", usersCollection, "index", "uniq_email")
	return &mongoUserRepository{coll: coll}, nil
}

func (r *mongoUserRepository) Create(ctx context.Context, user *User) error {
	doc := userDoc{
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		Plan:         user.Plan,
		CreatedAt:    user.CreatedAt,
	}

	res, err := r.coll.InsertOne(ctx, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrEmailTaken
		}
		return fmt.Errorf("auth: error insertando usuario: %w", err)
	}

	// Devolvemos el _id generado por Mongo como hex string al modelo de dominio.
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		user.ID = oid.Hex()
	}
	return nil
}

func (r *mongoUserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var doc userDoc
	err := r.coll.FindOne(ctx, bson.M{"email": email}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("auth: error buscando usuario por email: %w", err)
	}
	return doc.toDomain(), nil
}

func (r *mongoUserRepository) UpdatePlan(ctx context.Context, id string, plan Plan) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrUserNotFound
	}
	res, err := r.coll.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": bson.M{"plan": plan}})
	if err != nil {
		return fmt.Errorf("auth: error actualizando el plan del usuario: %w", err)
	}
	if res.MatchedCount == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *mongoUserRepository) UpdatePassword(ctx context.Context, id, passwordHash string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrUserNotFound
	}
	res, err := r.coll.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": bson.M{"passwordHash": passwordHash}})
	if err != nil {
		return fmt.Errorf("auth: error actualizando la password del usuario: %w", err)
	}
	if res.MatchedCount == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *mongoUserRepository) Delete(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrUserNotFound
	}
	res, err := r.coll.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		return fmt.Errorf("auth: error eliminando usuario: %w", err)
	}
	if res.DeletedCount == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *mongoUserRepository) FindByID(ctx context.Context, id string) (*User, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		// Un id con formato inválido no puede existir: lo tratamos como no encontrado.
		return nil, ErrUserNotFound
	}

	var doc userDoc
	err = r.coll.FindOne(ctx, bson.M{"_id": oid}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("auth: error buscando usuario por id: %w", err)
	}
	return doc.toDomain(), nil
}
