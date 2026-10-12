package skill

// Servicer defines the skill download contract used by the handler.
type Servicer interface {
	DownloadFoodLog(userID int) (*Download, error)
}
