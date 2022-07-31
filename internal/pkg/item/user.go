package item

type User struct {
	Username string          `firestore:"username,omitempty" json:"username,omitempty"`
	Image    string          `firestore:"image,omitempty" json:"image,omitempty"`
	Films    map[string]int  `firestore:"scores,omitempty" json:"scores,omitempty"`
	Songs    map[string]Song `firestore:"comments,omitempty" json:"comments,omitempty"`
}
