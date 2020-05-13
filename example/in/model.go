package in

type User struct{}

type EventSubForm struct {
	ID string

	Caption string

	Rank int32

	Fields *ArrayOfEventField

	User User

	PrimitivePointer *int

	SliceInt []int
}

type ArrayOfEventField struct {
	EventField []*EventField
}

type EventField struct {
	ID string `json:"id"`

	Name string

	FieldType string

	IsMandatory bool

	Rank int32

	Tag string

	Items *ArrayOfEventFieldItem

	CustomFieldOrder int32

	FieldDoesntExist string `elasticsearch:"no_source"`

	EmbeddedStruct
}

type EmbeddedStruct struct {
	DoubleEmbeddedStruct
	IdEmbedded int32
}

type DoubleEmbeddedStruct struct {
	IdDoubleEmbedded int32
}

type ArrayOfEventFieldItem struct {
	EventFieldItem []*EventFieldItem
}

type EventFieldItem struct {
	EventFieldItemID string

	Text string

	Rank int32

	FloatField1 float32
	FloatField2 float64
}

type Entity struct {
	EntityID string
	EmbeddedEntity
	SubEntities []SubEntity
}

type EmbeddedEntity struct {
	EmbeddedEntityID string
}

type SubEntity struct {
	SubEntityID string
}
