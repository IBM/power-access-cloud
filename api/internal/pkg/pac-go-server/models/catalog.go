package models

type Catalog struct {
	ID                      string        `json:"id"`
	Type                    string        `json:"type"`
	Name                    string        `json:"name"`
	Description             string        `json:"description"`
	Capacity                Capacity      `json:"capacity"`
	Retired                 bool          `json:"retired"`
	Expiry                  int           `json:"expiry"`
	ImageThumbnailReference string        `json:"image_thumbnail_reference"`
	VM                      VM            `json:"vm"`
	Status                  CatalogStatus `json:"status"`
}

type CatalogStatus struct {
	Ready   bool   `json:"ready"`
	Message string `json:"message,omitempty"`
}

type VM struct {
	CRN           string   `json:"crn"`
	ProcessorType string   `json:"processor_type"`
	SystemType    string   `json:"system_type"`
	Image         string   `json:"image"`
	Network       string   `json:"network"`
	Capacity      Capacity `json:"capacity"`
}
