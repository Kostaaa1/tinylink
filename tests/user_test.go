package user_test

// func TestUser(t *testing.T) {
// 	err := exec.Command("rm", "-rf", "./tinylink.db").Run()
// 	require.Nil(t, err)

// 	conf := config.SQLConfig{
// 		SQLitePath:   "./tinylink.db",
// 		MaxOpenConns: 25,
// 		MaxIdleConns: 25,
// 	}

// 	db, err := sqlitedb.StartDB(conf)
// 	require.Nil(t, err)

// 	provider := user.NewRepositoryProvider(db)
// 	userService := user.NewService(provider)

// 	ctx := context.Background()

// 	gUser := &user.GoogleUser{
// 		ID:            "33333",
// 		Email:         "kostaarsic123@gmail.com",
// 		Name:          "Kosta",
// 		GivenName:     "Arsic",
// 		FamilyName:    "Kosta Arsic",
// 		Picture:       "testpicture-url.com",
// 		VerifiedEmail: true,
// 	}

// 	userDTO, err := userService.HandleGoogleLogin(ctx, gUser)
// 	require.Nil(t, err)
// 	require.NotZero(t, userDTO.CreatedAt)
// 	require.NotNil(t, userDTO)

// 	newUser := &user.User{
// 		Email: userDTO.Email,
// 		Name:  userDTO.Name,
// 	}
// 	err = newUser.Password.Set("TestPassword123")
// 	require.Nil(t, err)

// 	err = userService.Register(ctx, newUser)
// 	require.Nil(t, err)

// 	userRepo := provider.GetDbAdapters().UserRepository
// 	fetchedUser, err := userRepo.GetByEmail(ctx, userDTO.Email)
// 	require.Nil(t, err)

// 	require.NotEmpty(t, fetchedUser)
// 	require.NotNil(t, fetchedUser)

// 	t.Log("logovan: ", fetchedUser)
// }
