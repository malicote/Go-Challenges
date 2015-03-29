This is my solution to Go challenge #1:
http://golang-challenge.com/go-challenge1/

It was a straight-forward binary parsing challenge to decode drum machine files.  
I read drum-tabs endlessly as a teenager, so this was a great challenge to solve.  
So much so that I learned Go to do so.  I had never written Go and had to use Google's
introduction tutorials and install Go to solve the challenge.


**decoder.go** provides the two data structures and the interface for testing.  
The main data structure is *pattern*, which holds all the data for the sample.  Each track
is stored in a *track* struct.

**drum.go** provides the parsing and formatted printing functionality.

The testing infrastructure can be found in the Go challenge website.  For copyright purposes
I will not post it here.

The binary data is stored in the following format (listed as a c-style structs):
```
struct splice {
  char          splice_magic_cookie[6]  // SPLICE
  uint64_t      length_bytes            // Length of rest of data (big endian)
  char          hardware_rev[32]
  float         tempo                   // Little endian
  struct track  tracks[];
};
//  --- repeat for each track until read == (length_bytes - 36) --:
struct track {
  uint8_t     track_name_length
  char        track_name[track_name_length]
  uint8_t     steps[16]
};
```
