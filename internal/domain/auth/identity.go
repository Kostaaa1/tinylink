package auth

// type Identity struct {
// 	IsAuth bool
// 	UserID *uint64
// 	guestID *string
// }

// func GetIdentifiers(r *http.Request) Identity {
// 	var identity Identity

// 	claims, _ := ClaimsFromRequest(r)
// 	sessionID := GetguestID(r)

// 	if claims.UserID > 0 {
// 		identity.IsAuth = true
// 		identity.UserID = &claims.UserID
// 	} else if sessionID != nil {
// 		identity.guestID = sessionID
// 	}

// 	return identity
// }
