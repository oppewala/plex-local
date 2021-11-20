package plex

type Directory struct {
	AllowSync        bool       `json:"allowSync"`
	Art              string     `json:"art"`
	Composite        string     `json:"composite"`
	Filters          bool       `json:"filters"`
	Refreshing       bool       `json:"refreshing"`
	Thumb            string     `json:"thumb"`
	Key              string     `json:"key"`
	Type             string     `json:"type"`
	Title            string     `json:"title"`
	Agent            string     `json:"agent"`
	Scanner          string     `json:"scanner"`
	Language         string     `json:"language"`
	UUID             string     `json:"uuid"`
	UpdatedAt        int        `json:"updatedAt"`
	CreatedAt        int        `json:"createdAt"`
	ScannedAt        int        `json:"scannedAt"`
	Content          bool       `json:"content"`
	Directory        bool       `json:"directory"`
	ContentChangedAt int        `json:"contentChangedAt"`
	Hidden           int        `json:"hidden"`
	Location         []Location `json:"Location"`
}
type LibrariesMediaContainer struct {
	Size            int         `json:"size"`
	AllowSync       bool        `json:"allowSync"`
	Identifier      string      `json:"identifier"`
	MediaTagPrefix  string      `json:"mediaTagPrefix"`
	MediaTagVersion int         `json:"mediaTagVersion"`
	Title1          string      `json:"title1"`
	Directory       []Directory `json:"Directory"`
}
type LibrariesRoot struct {
	MediaContainer LibrariesMediaContainer `json:"MediaContainer"`
}

type Location struct {
	ID   int    `json:"id"`
	Path string `json:"path"`
}
type Director struct {
	Tag string `json:"tag"`
}
type Country struct {
	Tag string `json:"tag"`
}
type Genre struct {
	Tag string `json:"tag"`
}
type Role struct {
	Tag string `json:"tag"`
}
type Part struct {
	ID           int    `json:"id"`
	Key          string `json:"key"`
	Duration     int    `json:"duration"`
	File         string `json:"file"`
	Size         uint64 `json:"size"`
	Container    string `json:"container"`
	VideoProfile string `json:"videoProfile"`
}
type Media struct {
	ID              int     `json:"id"`
	Duration        int     `json:"duration"`
	Bitrate         int     `json:"bitrate"`
	Width           int     `json:"width"`
	Height          int     `json:"height"`
	AspectRatio     float64 `json:"aspectRatio"`
	AudioChannels   int     `json:"audioChannels"`
	AudioCodec      string  `json:"audioCodec"`
	VideoCodec      string  `json:"videoCodec"`
	VideoResolution string  `json:"videoResolution"`
	Container       string  `json:"container"`
	VideoFrameRate  string  `json:"videoFrameRate"`
	VideoProfile    string  `json:"videoProfile"`
	Part            []Part  `json:"Part"`
}
type Writer struct {
	Tag string `json:"tag"`
}
type GUID struct {
	ID string `json:"id"`
}
type Metadata struct {
	RatingKey             string     `json:"ratingKey"`
	Key                   string     `json:"key"`
	ParentRatingKey       string     `json:"parentRatingKey"`
	GrandparentRatingKey  string     `json:"grandparentRatingKey"`
	Guid                  string     `json:"guid"` // Some media returns 2 guid properties...
	GUID                  []GUID     `json:"Guid"`
	Studio                string     `json:"studio"`
	Type                  string     `json:"type"`
	Title                 string     `json:"title"`
	GrandparentKey        string     `json:"grandparentKey"`
	ParentKey             string     `json:"parentKey"`
	GrandparentTitle      string     `json:"grandparentTitle"`
	ParentTitle           string     `json:"parentTitle"`
	ContentRating         string     `json:"contentRating"`
	Summary               string     `json:"summary"`
	AudienceRating        float64    `json:"audienceRating,omitempty"`
	Index                 int        `json:"index"`
	ParentIndex           int        `json:"parentIndex"`
	Rating                float64    `json:"rating,omitempty"`
	ViewCount             int        `json:"viewCount,omitempty"`
	LastViewedAt          int        `json:"lastViewedAt,omitempty"`
	Year                  int        `json:"year,omitempty"`
	Thumb                 string     `json:"thumb"`
	Art                   string     `json:"art"`
	ParentThumb           string     `json:"parentThumb"`
	GrandparentThumb      string     `json:"grandparentThumb"`
	GrandparentArt        string     `json:"grandparentArt"`
	GrandparentTheme      string     `json:"grandparentTheme"`
	Duration              int        `json:"duration"`
	OriginallyAvailableAt string     `json:"originallyAvailableAt,omitempty"`
	AddedAt               int        `json:"addedAt"`
	UpdatedAt             int        `json:"updatedAt"`
	Media                 []Media    `json:"Media"`
	Genre                 []Genre    `json:"Genre,omitempty"`
	Director              []Director `json:"Director,omitempty"`
	Writer                []Writer   `json:"Writer,omitempty"`
	Country               []Country  `json:"Country,omitempty"`
	Role                  []Role     `json:"Role,omitempty"`
	TitleSort             string     `json:"titleSort,omitempty"`
	OriginalTitle         string     `json:"originalTitle,omitempty"`
	ViewOffset            int        `json:"viewOffset,omitempty"`
	AudienceRatingImage   string     `json:"audienceRatingImage,omitempty"`
	PrimaryExtraKey       string     `json:"primaryExtraKey,omitempty"`
	Tagline               string     `json:"tagline,omitempty"`
	Banner                string     `json:"banner,omitempty"`
	LeafCount             int        `json:"leafCount"`
	ViewedLeafCount       int        `json:"viewedLeafCount"`
	ChildCount            int        `json:"childCount"`
	Theme                 string     `json:"theme,omitempty"`
	SkipCount             int        `json:"skipCount,omitempty"`
	UserRating            float64    `json:"userRating,omitempty"`
	//Guid []GUID `json:"GUID,omitempty"` // Some media returns 2 guid properties...
}
type MediaContainer struct {
	Size                int         `json:"size"`
	AllowSync           bool        `json:"allowSync"`
	Art                 string      `json:"art"`
	Banner              string      `json:"banner"`
	Identifier          string      `json:"identifier"`
	Key                 string      `json:"key"`
	LibrarySectionID    int         `json:"librarySectionID"`
	LibrarySectionTitle string      `json:"librarySectionTitle"`
	LibrarySectionUUID  string      `json:"librarySectionUUID"`
	MediaTagPrefix      string      `json:"mediaTagPrefix"`
	MediaTagVersion     int         `json:"mediaTagVersion"`
	MixedParents        bool        `json:"mixedParents"`
	Nocache             bool        `json:"nocache"`
	ParentIndex         int         `json:"parentIndex"`
	ParentTitle         string      `json:"parentTitle"`
	ParentYear          int         `json:"parentYear"`
	Thumb               string      `json:"thumb"`
	Theme               string      `json:"theme"`
	Title1              string      `json:"title1"`
	Title2              string      `json:"title2"`
	ViewGroup           string      `json:"viewGroup"`
	ViewMode            int         `json:"viewMode"`
	Metadata            []Metadata  `json:"Metadata"`
	Directory           []Directory `json:"Directory"`
}
type ResponseRoot struct {
	MediaContainer MediaContainer `json:"MediaContainer"`
}
