package user

import (
	"backend/internal/models"
	"backend/internal/repository"
	"backend/internal/service"
	"backend/pkg/cerr"
	"backend/pkg/log"
	"context"
	"golang.org/x/crypto/bcrypt"
)

type ServUser struct {
	UserRepo repository.UserRepo
}

func InitUserService(userRepo repository.UserRepo) service.UserServ {
	return &ServUser{UserRepo: userRepo}
}

func (s ServUser) Create(ctx context.Context, user models.UserCreate) (int, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PWD), 10)
	if err != nil {
		log.Log.Error(err)
		return 0, err
	}
	newUser := models.UserCreate{
		UserBase: user.UserBase,
		PWD:      string(hashedPassword),
	}
	id, err := s.UserRepo.Create(ctx, newUser)
	if err != nil {
		log.Log.Error(err)
		return 0, err
	}
	return id, nil
}

func (s ServUser) Get(ctx context.Context, id int) (*models.User, error) {
	user, err := s.UserRepo.Get(ctx, id)
	if err != nil {
		log.Log.Error(err)
		return nil, err
	}
	return user, nil
}

func (s ServUser) GetAll(ctx context.Context) ([]models.User, error) {
	users, err := s.UserRepo.GetAll(ctx)
	if err != nil {
		log.Log.Error(err)
		return nil, err
	}
	return users, nil
}

func (s ServUser) Login(ctx context.Context, user models.UserLogin) (int, error) {
	id, pwd, err := s.UserRepo.GetPWDbyEmail(ctx, user.Email)
	if err != nil {
		log.Log.Error(err)
		return 0, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(pwd), []byte(user.PWD))
	if err != nil {
		log.Log.Error(cerr.InvalidPWD(err))
		return 0, cerr.InvalidPWD(err)
	}
	return id, nil
}

func (s ServUser) ChangePWD(ctx context.Context, user models.UserChangePWD) (int, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(user.NewPWD), 10)
	if err != nil {

		log.Log.Error(cerr.Hash(err))
		return 0, cerr.Hash(err)
	}
	newPWD := models.UserChangePWD{
		ID:     user.ID,
		NewPWD: string(hash),
	}
	id, err := s.UserRepo.ChangePWD(ctx, newPWD)
	if err != nil {
		log.Log.Error(err)
		return 0, err
	}
	return id, nil
}

func (s ServUser) Delete(ctx context.Context, id int) error {
	err := s.UserRepo.Delete(ctx, id)
	if err != nil {
		log.Log.Error(err)
		return err
	}
	return nil
}
