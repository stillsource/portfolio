package domain

// Poetry holds the optional literary content attached to a Roll.
//
// GlobalPoem is the markdown body that applies to the whole Roll.
// PhotoPoems maps a photo file name to its dedicated verse.
type Poetry struct {
	GlobalPoem string
	PhotoPoems map[string]string
}
