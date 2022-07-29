package item

type UserID string

type User struct {
	Username string          `firestore:"username,omitempty" json:"username,omitempty"`
	Image    string          `firestore:"image,omitempty" json:"image,omitempty"`
	Films    map[FilmID]int  `firestore:"scores,omitempty" json:"scores,omitempty"`
	Songs    map[SongID]Song `firestore:"comments,omitempty" json:"comments,omitempty"`
}
