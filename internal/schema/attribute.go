package schema

type AttributeType string

const (
	AttributeTypeEmpty  AttributeType = "Empty"
	AttributeTypeStr    AttributeType = "Str"
	AttributeTypeBool   AttributeType = "Bool"
	AttributeTypeInt    AttributeType = "Int"
	AttributeTypeDouble AttributeType = "Double"
	AttributeTypeMap    AttributeType = "Map"
	AttributeTypeSlice  AttributeType = "Slice"
	AttributeTypeBytes  AttributeType = "Bytes"
)

type RequirementLevel string

const (
	RequirementLevelRequired    RequirementLevel = "Required"
	RequirementLevelRecommended RequirementLevel = "Recommended"
	RequirementLevelOptIn       RequirementLevel = "OptIn"
)

type AttributeSource string

const (
	AttributeSourceResource AttributeSource = "Resource"
	AttributeSourceScope    AttributeSource = "Scope"

	AttributeSourceDataPoint AttributeSource = "DataPoint"
	AttributeSourceLogRecord AttributeSource = "LogRecord"
	AttributeSourceSpan      AttributeSource = "Span"
	AttributeSourceProfile   AttributeSource = "Profile"
)

type Attribute struct {
	// Name is the name of the attribute.
	Name string `json:"name"`

	// Type is the type of the attribute.
	Type AttributeType `json:"type"`

	// Brief is a brief description of the attribute.
	Brief string `json:"brief,omitempty"`
	// Examples is a sequence of example values for the attribute or single example
	// value. They are required only for string and string array
	// attributes. Example values must be of the same type of the
	// attribute. If only a single example is provided, it can directly
	// be reported without encapsulating it into a sequence/dictionary.
	Examples *interface{} `json:"examples,omitempty"`

	// Specifies if the attribute is mandatory. Can be "required",
	// "conditionally_required", "recommended" or "opt_in". When omitted,
	// the attribute is "recommended". When set to
	// "conditionally_required", the string provided as <condition> MUST
	// specify the conditions under which the attribute is required.
	RequirementLevel RequirementLevel `json:"requirement_level,omitempty"`

	// Source is the source of the attribute.
	// If the attribute is a resource attribute, the source is "Resource".
	// If the attribute is a scope attribute, the source is "Scope".
	// If the attribute is a data point attribute, the source is "DataPoint".
	Source AttributeSource `json:"source,omitempty"`
}
