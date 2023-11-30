package requests

type SignupRequest struct {
	Login     string `json:"login"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	BirthDate string `json:"birth_date"`
	Name      string `json:"name"`
}

type SigninRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type CommentRequest struct {
	FilmId uint64 `json:"film_id"`
	Rating uint16 `json:"rating"`
	Text   string `json:"text"`
}

type EditProfileRequest struct {
	Login    string `json:"login"`
	Email    string `json:"email"`
	Photo    []byte `json:"photo"`
	Password string `json:"password"`
}

type FindFilmRequest struct {
	Title      string   `json:"title"`
	DateFrom   string   `json:"date_from"`
	DateTo     string   `json:"date_to"`
	RatingFrom float32  `json:"rating_from"`
	RatingTo   float32  `json:"rating_to"`
	Mpaa       string   `json:"mpaa"`
	Genres     []string `json:"genres"`
	Actors     []string `json:"actors"`
}
