package repository

import (
	"context"
	"errors"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"api-students/app/model"
)

var (
	ErrPortfolioNotFound = errors.New("portofolio tidak ditemukan")
	ErrProjectLimit      = errors.New("batas maksimal 20 proyek tercapai")
)

const MaxProjectsPerPortfolio = 20

type PortfolioRepository interface {
	EnsureIndexes(ctx context.Context) error
	FindByStudentID(ctx context.Context, studentID int) (model.Portfolio, error)
	Upsert(ctx context.Context, p model.Portfolio) (model.Portfolio, error)
	AddProject(ctx context.Context, studentID int, project model.Project) error
	DeleteProject(ctx context.Context, studentID int, projectID string) error
	GetTopTechnologies(ctx context.Context) ([]bson.M, error)
}

type portfolioMongoRepository struct {
	coll *mongo.Collection
}

func NewPortfolioRepository(db *mongo.Database) PortfolioRepository {
	return &portfolioMongoRepository{coll: db.Collection("student_portfolios")}
}

// 1. Index Unik untuk StudentID (Satu mahasiswa 1 portofolio)
func (r *portfolioMongoRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "student_id", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("student_unique"),
	})
	return err
}

// 2. GET /students/:id/portfolio
func (r *portfolioMongoRepository) FindByStudentID(ctx context.Context, studentID int) (model.Portfolio, error) {
	var p model.Portfolio
	err := r.coll.FindOne(ctx, bson.M{"student_id": studentID}).Decode(&p)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return model.Portfolio{}, ErrPortfolioNotFound
	}
	return p, err
}

// 3. PUT /students/:id/portfolio (Upsert)
func (r *portfolioMongoRepository) Upsert(ctx context.Context, p model.Portfolio) (model.Portfolio, error) {
	now := time.Now().UTC().Truncate(time.Millisecond)
	filter := bson.M{"student_id": p.StudentID}
	update := bson.M{
		"$set": bson.M{
			"headline":   p.Headline,
			"bio":        p.Bio,
			"updated_at": now,
		},
		"$setOnInsert": bson.M{
			"projects":     []model.Project{},
			"certificates": []model.Certificate{},
			"created_at":   now,
		},
	}
	
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	var result model.Portfolio
	err := r.coll.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	return result, err
}

// 4. POST /students/:id/portfolio/projects (Atomic Push + Max Limit)
func (r *portfolioMongoRepository) AddProject(ctx context.Context, studentID int, project model.Project) error {
	project.ID = bson.NewObjectID().Hex()
	
	// Trik batas ukuran: jika element ke-20 (index 19) BELUM ada, maka boleh push
	limitField := "projects." + strconv.Itoa(MaxProjectsPerPortfolio-1)
	
	filter := bson.M{
		"student_id": studentID,
		limitField:   bson.M{"$exists": false}, 
	}
	update := bson.M{
		"$push": bson.M{"projects": project},
		"$set":  bson.M{"updated_at": time.Now().UTC()},
	}

	res, err := r.coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		// Cek apakah gagal karena portofolionya tidak ada atau karena limit
		count, _ := r.coll.CountDocuments(ctx, bson.M{"student_id": studentID})
		if count == 0 {
			return ErrPortfolioNotFound
		}
		return ErrProjectLimit
	}
	return nil
}

// 5. DELETE /students/:id/portfolio/projects/:projectId
func (r *portfolioMongoRepository) DeleteProject(ctx context.Context, studentID int, projectID string) error {
	filter := bson.M{"student_id": studentID}
	update := bson.M{
		"$pull": bson.M{"projects": bson.M{"id": projectID}},
		"$set":  bson.M{"updated_at": time.Now().UTC()},
	}
	res, err := r.coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return ErrPortfolioNotFound // Atau project tidak ditemukan
	}
	return nil
}

// 6. GET /portfolios/stats/technologies (Aggregation $unwind 2x)
func (r *portfolioMongoRepository) GetTopTechnologies(ctx context.Context) ([]bson.M, error) {
	// 1. Unwind array projects
	// 2. Unwind array techs di dalam projects
	// 3. Group by tech, hitung total
	// 4. Sort descending, ambil 10
	pipeline := mongo.Pipeline{
		{{Key: "$unwind", Value: "$projects"}},
		{{Key: "$unwind", Value: "$projects.techs"}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$projects.techs",
			"total": bson.M{"$sum": 1},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "total", Value: -1}}}},
		{{Key: "$limit", Value: 10}},
	}

	cursor, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}