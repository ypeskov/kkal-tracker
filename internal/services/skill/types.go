package skill

// Download is a skill packed as a zip, ready to be installed in an agent
type Download struct {
	FileName string
	Content  []byte
}
