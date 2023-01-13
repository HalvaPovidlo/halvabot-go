package item

type User struct {
	ID       string `firestore:"-" json:"id"`
	Username string `firestore:"username" json:"username,omitempty"`
	Avatar   string `firestore:"avatar,omitempty" json:"avatar,omitempty"`
}
