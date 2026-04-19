package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"practice8/repository"
)

func TestGetUserByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	mockRepo := repository.NewMockUserRepository(ctrl)
	userService := NewUserService(mockRepo)
	user := &repository.User{ID: 1, Name: "Bakytzhan Agai"}
	mockRepo.EXPECT().GetUserByID(1).Return(user, nil)
	result, err := userService.GetUserByID(1)
	assert.NoError(t, err)
	assert.Equal(t, user, result)
}

func TestCreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	mockRepo := repository.NewMockUserRepository(ctrl)
	userService := NewUserService(mockRepo)
	user := &repository.User{ID: 1, Name: "Bakytzhan Agai"}
	mockRepo.EXPECT().CreateUser(user).Return(nil)
	err := userService.CreateUser(user)
	assert.NoError(t, err)
}

func TestRegisterUser_UserAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	mockRepo := repository.NewMockUserRepository(ctrl)
	svc := NewUserService(mockRepo)
	existing := &repository.User{ID: 2, Name: "Other", Email: "a@b.c"}
	mockRepo.EXPECT().GetByEmail("a@b.c").Return(existing, nil)
	err := svc.RegisterUser(&repository.User{Name: "New"}, "a@b.c")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestRegisterUser_NewUserSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	mockRepo := repository.NewMockUserRepository(ctrl)
	svc := NewUserService(mockRepo)
	u := &repository.User{ID: 1, Name: "Bakytzhan Agai", Email: "new@example.com"}
	mockRepo.EXPECT().GetByEmail("new@example.com").Return(nil, nil)
	mockRepo.EXPECT().CreateUser(u).Return(nil)
	err := svc.RegisterUser(u, "new@example.com")
	assert.NoError(t, err)
}

func TestRegisterUser_GetByEmailRepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	mockRepo := repository.NewMockUserRepository(ctrl)
	svc := NewUserService(mockRepo)
	mockRepo.EXPECT().GetByEmail("x@y.z").Return(nil, errors.New("db down"))
	err := svc.RegisterUser(&repository.User{Name: "X"}, "x@y.z")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error getting user with this email")
}

func TestRegisterUser_CreateUserRepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	mockRepo := repository.NewMockUserRepository(ctrl)
	svc := NewUserService(mockRepo)
	u := &repository.User{Name: "N"}
	mockRepo.EXPECT().GetByEmail("n@n.n").Return(nil, nil)
	mockRepo.EXPECT().CreateUser(u).Return(errors.New("insert failed"))
	err := svc.RegisterUser(u, "n@n.n")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "insert failed")
}

func TestUpdateUserName_EmptyName(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	svc := NewUserService(repository.NewMockUserRepository(ctrl))
	err := svc.UpdateUserName(10, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name cannot be empty")
}

func TestUpdateUserName_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	mockRepo := repository.NewMockUserRepository(ctrl)
	svc := NewUserService(mockRepo)
	mockRepo.EXPECT().GetUserByID(99).Return(nil, errors.New("not found"))
	err := svc.UpdateUserName(99, "New")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestUpdateUserName_Success_NameChanged(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	mockRepo := repository.NewMockUserRepository(ctrl)
	svc := NewUserService(mockRepo)
	u := &repository.User{ID: 2, Name: "Old"}
	mockRepo.EXPECT().GetUserByID(2).Return(u, nil)
	mockRepo.EXPECT().UpdateUser(gomock.AssignableToTypeOf(&repository.User{})).DoAndReturn(func(user *repository.User) error {
		assert.Equal(t, "New", user.Name)
		assert.Equal(t, 2, user.ID)
		return nil
	})
	err := svc.UpdateUserName(2, "New")
	assert.NoError(t, err)
}

func TestUpdateUserName_SameName_SkipsUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	mockRepo := repository.NewMockUserRepository(ctrl)
	svc := NewUserService(mockRepo)
	u := &repository.User{ID: 2, Name: "Same"}
	mockRepo.EXPECT().GetUserByID(2).Return(u, nil)
	err := svc.UpdateUserName(2, "Same")
	assert.NoError(t, err)
}

func TestUpdateUserName_UpdateUserFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	mockRepo := repository.NewMockUserRepository(ctrl)
	svc := NewUserService(mockRepo)
	u := &repository.User{ID: 3, Name: "Old"}
	mockRepo.EXPECT().GetUserByID(3).Return(u, nil)
	mockRepo.EXPECT().UpdateUser(gomock.AssignableToTypeOf(&repository.User{})).Return(errors.New("constraint"))
	err := svc.UpdateUserName(3, "X")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "constraint")
}

func TestDeleteUser_AdminForbidden(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	svc := NewUserService(repository.NewMockUserRepository(ctrl))
	err := svc.DeleteUser(1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not allowed to delete admin")
}

func TestDeleteUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	mockRepo := repository.NewMockUserRepository(ctrl)
	svc := NewUserService(mockRepo)
	mockRepo.EXPECT().DeleteUser(42).Return(nil)
	err := svc.DeleteUser(42)
	assert.NoError(t, err)
}

func TestDeleteUser_RepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	mockRepo := repository.NewMockUserRepository(ctrl)
	svc := NewUserService(mockRepo)
	mockRepo.EXPECT().DeleteUser(7).Return(errors.New("db error"))
	err := svc.DeleteUser(7)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}
