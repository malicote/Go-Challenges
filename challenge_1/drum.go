// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import (
    "bytes"
    "encoding/binary"
    "errors"
    "fmt"
    "strings"
)


// Formated printing functions

// Print to buffer to make it more efficient.
func print_track_to_buffer(track Track, buffer *bytes.Buffer) {
    // Set up the formatting, this could be somewhere else
    tab_format := [2]string{"-" , "x"}

    // Print the track
    buffer.WriteString(fmt.Sprintf("(%d) %s\t", track.ID, track.name))
    for count, step := range track.steps {
        if count % 4 == 0 {
            buffer.WriteString(fmt.Sprintf("|"))
        }

        buffer.WriteString(fmt.Sprintf("%s", tab_format[step]))
    }
    buffer.WriteString(fmt.Sprintf("|\n"))
}

// Prints entire pattern
func print_pattern(pattern Pattern) string {
    var buffer bytes.Buffer
    buffer.WriteString(fmt.Sprintf("Saved with HW Version: %s\n", pattern.hardware_rev))
    buffer.WriteString(fmt.Sprintf("Tempo: %g\n", pattern.tempo))
    for _, track := range pattern.tracks {
        print_track_to_buffer(track, &buffer)
    }

    return buffer.String()
}

// Binary parsing functions

// Basic binary format (C-style struct):
//  char[6]     splice_magic_cookie // SPLICE
//  uint64_t    length_bytes        // Length of rest of data (big endian)
//  char[32]    hardware_rev
//  float       tempo               // Little endian
//  --- repeat for each track until read == (length_bytes - 36) --:
//  uint8_t                 track_name_length
//  char[track_name_length] track_name
//  uint8_t[16]             steps
//  ---------------------------------------------------------------

// Use offsets since this is all slicing
type BinaryPatternData struct {
    magic_cookie []byte
    // Binary offset for pattern
    magic_cookie_start int
    magic_cookie_end int
    binary_length_start int
    binary_length_end int
    hardware_rev_start int
    hardware_rev_end int
    tempo_start int
    tempo_end int
    track_start int
    // Track data, offsets are relative
    ID_start int
    ID_end int
    name_length_start int
    name_length_end int
    step_length int
    // track name is variable length
    // steps start after track name
}

// Parses a given binary into the Pattern struct
func parse_splice_block(data []byte) (Pattern, error) {
    var pattern Pattern

    // TODO: make this more generic/auto-filling.
    pattern_offsets := BinaryPatternData {
        magic_cookie: []byte{'S', 'P', 'L', 'I', 'C', 'E'},
        magic_cookie_start: 0,
        magic_cookie_end: 6,
        binary_length_start: 6,
        binary_length_end: 14,
        hardware_rev_start: 14,
        hardware_rev_end: 46,
        tempo_start: 46,
        tempo_end: 50,
        track_start: 50,
        ID_start: 0,
        ID_end: 4,
        name_length_start: 4,
        name_length_end: 5,
        step_length: 16,
    }

    // Would be nice to make all of this generic

    if !valid_magic_cookie(data, pattern_offsets) {
        return pattern, errors.New("Not a splice file!")
    }

    max_length, err := get_binary_data_length(data, pattern_offsets)
    if err != nil {
        return pattern, err
    }

    // Make sure we have enough data left
    // Why len() returns int I'll never know!
    data_left := uint64(len(data[pattern_offsets.binary_length_end:]))
    if data_left < max_length {
        return pattern, errors.New("Data is missing information!\n")
    }

    pattern.hardware_rev, err = get_hardware_rev(data, pattern_offsets)
    if err != nil {
        return pattern, err
    }

    pattern.tempo, err = get_tempo(data, pattern_offsets)
    if err != nil {
        return pattern, err
    }

    // max_length includes some of the header data
    track_data_len := int(max_length) -
                      (pattern_offsets.track_start - pattern_offsets.binary_length_end)

    track_data_start := pattern_offsets.track_start
    track_data_end := track_data_start + track_data_len + 1
    track_data := data[track_data_start:track_data_end]

    // Iterate until no data is left (compared to max_length)
    var total_read = 0

    // This cast should be OK for splice purposes
    for total_read < int(track_data_len) {
        var track Track
        track, read, err := parse_track(track_data[total_read:], pattern_offsets)
        if err != nil {
            return pattern, err
        }

        total_read += read
        pattern.tracks = append(pattern.tracks, track)
    }
    return pattern, err
}

// Checks that SPLICE is the first 6 bytes
func valid_magic_cookie(data []byte, offsets BinaryPatternData) bool {
    magic_cookie := offsets.magic_cookie
    start := offsets.magic_cookie_start
    end := offsets.magic_cookie_end

    return len(data) > end && bytes.Compare(magic_cookie, data[start:end]) == 0
}

func get_binary_data_length(data []byte, offsets BinaryPatternData) (uint64, error) {
    start := offsets.binary_length_start
    end :=  offsets.binary_length_end

    if len(data) < end {
        return 0, errors.New("Data not long enough to provide data length\n")
    }

    var binary_length uint64 = 0
    buf := bytes.NewReader(data[start:end])
    err := binary.Read(buf, binary.BigEndian, &binary_length)

    return binary_length, err
}

func get_hardware_rev(data []byte, offsets BinaryPatternData) (string, error) {
    start := offsets.hardware_rev_start
    end := offsets.hardware_rev_end

    if len(data) < end {
        return "", errors.New("Data not long enough to provide hardware rev\n")
    }

    // Remove the extra 0's from the data string, since Go doesn't need them
    return strings.Trim(string(data[start:end]), "\x00"), nil
}

func get_tempo(data []byte, offsets BinaryPatternData) (float32, error) {
    start := offsets.tempo_start
    end := offsets.tempo_end

    if len(data) < end {
        return 0.0, errors.New("Data is not long enough to provide tempo!\n")
    }

    var tempo float32 = 0
    buf := bytes.NewReader(data[start:end])
    err := binary.Read(buf, binary.LittleEndian, &tempo)

    return tempo, err
}

// Returns error, Track, and how many bytes were consumed while reading
// This is probably a good candidate for the io streaming features of go
func parse_track(data []byte, offsets BinaryPatternData) (Track, int, error) {
    var track Track

    if len(data) < offsets.ID_end {
        return track, 0, errors.New("Data not big enough to hold track info\n")
    }

    // Get ID
    start := offsets.ID_start
    end := offsets.ID_end

    buf := bytes.NewReader(data[start:end])
    err := binary.Read(buf, binary.LittleEndian, &track.ID)
    if err != nil {
        return track, offsets.ID_end, err
    }

    // This should only take one line?
    var track_name_length uint8
    start = offsets.name_length_start
    end = offsets.name_length_end

    buf = bytes.NewReader(data[start:end])
    err = binary.Read(buf, binary.LittleEndian, &track_name_length)
    if err != nil {
        return track, offsets.name_length_end, err
    }

    // Get name
    start = offsets.name_length_end
    end = offsets.name_length_end + int(track_name_length)
    track.name = string(data[start:end])

    // Get steps, use previous end calculation
    start = end
    end = start + offsets.step_length
    track.steps = append(track.steps, data[start:end]...)
    if len(track.steps) != offsets.step_length {
        return track,
               offsets.name_length_end + int(track_name_length) + len(track.steps),
               errors.New("Error reading track data.\n")
    }
    return track, offsets.name_length_end + int(track_name_length) + len(track.steps), nil
}
