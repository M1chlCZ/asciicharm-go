package ascii

type CharSet int

const (
	CharSetClassic CharSet = iota // your original 16-ish level ramp
	CharSetPhoto                  // long, smooth photographic ramp
	CharSetMinimal                // @%#*+=-:. style
	CharSetBlocks                 // " ░▒▓█" block characters
)

const (
	asciiClassic    = " .,:;i1tfLCG08@"
	asciiClassicInv = "@80GCLft1i;:,. "
)

// Photo-like, high-detail ramp (great with dithering)
const (
	asciiPhoto    = " .'`^\",:;Il!i><~+_-?][}{1)(|\\/*tfjrxnuvczXYUJCLQ0OZmwqpdbkhao*#MW&8%B@$"
	asciiPhotoInv = "@$B%8&WM#*ao bhkdpqwmZO0QLCJUYXzcvunxrjft/\\|)(1}{][?-_+~<>i!lI;:,'^`. "
)

// Minimal but nice for photos
const (
	asciiMinimal    = "@%#*+=-:. "
	asciiMinimalInv = " .:-=+*#%@"
)

// Block-style (needs a font that supports box-drawing)
const (
	asciiBlocks    = " ░▒▓█"
	asciiBlocksInv = "█▓▒░ "
)
