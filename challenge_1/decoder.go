package drum

import (
  "io/ioutil"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	  p := &Pattern{}

    dat, err := ioutil.ReadFile(path)
    if err != nil {
        return p, err
    }

    var pattern Pattern
    pattern, err = parse_splice_block(dat)
    if err != nil {
        return p, err
    }

    return &pattern, nil
}

// Holds data for an individual track
type Track struct {
    ID uint8
    name string
    steps []uint8
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
    hardware_rev string
    tempo float32
    tracks []Track
}

func (p Pattern) String() string {
   return print_pattern(p)
}
