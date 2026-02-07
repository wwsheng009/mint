package theme

// ColorValue represents a complete color definition with RGB, ANSI 16-color, and ANSI 256-color values
type ColorValue struct {
	RGB     [3]int // [R, G, B] for TrueColor
	ANSI16  int    // 0-15 for 16-color terminals
	ANSI256 int    // 0-255 for 256-color terminals
}

// ThemePreset defines a complete theme preset with all 23 semantic colors
type ThemePreset struct {
	Name   string
	Colors map[string]ColorValue
}

// Presets returns all available theme presets
func Presets() map[string]ThemePreset {
	return map[string]ThemePreset{
		"nord":            nordTheme(),
		"dracula":         draculaTheme(),
		"gruvbox-dark":    gruvboxDarkTheme(),
		"catppuccin-mocha": catppuccinMochaTheme(),
		"solarized-dark":  solarizedDarkTheme(),
	}
}

// nordTheme implements the Nord color scheme (arctic, north-bluish cleanliness)
func nordTheme() ThemePreset {
	return ThemePreset{
		Name: "nord",
		Colors: map[string]ColorValue{
			// Layer System
			"bg":      {RGB: [3]int{46, 52, 64}, ANSI16: 0, ANSI256: 235},
			"surface": {RGB: [3]int{59, 66, 82}, ANSI16: 8, ANSI256: 237},
			"overlay": {RGB: [3]int{40, 44, 52}, ANSI16: 0, ANSI256: 234},

			// Typography
			"text":        {RGB: [3]int{236, 239, 244}, ANSI16: 7, ANSI256: 253},
			"muted":       {RGB: [3]int{97, 110, 136}, ANSI16: 8, ANSI256: 244},
			"placeholder": {RGB: [3]int{97, 110, 136}, ANSI16: 8, ANSI256: 244},

			// Brand & Action
			"primary":   {RGB: [3]int{136, 192, 208}, ANSI16: 6, ANSI256: 110},
			"secondary": {RGB: [3]int{129, 161, 193}, ANSI16: 4, ANSI256: 109},
			"accent":    {RGB: [3]int{143, 188, 187}, ANSI16: 14, ANSI256: 117},

			// State
			"success": {RGB: [3]int{163, 190, 140}, ANSI16: 2, ANSI256: 108},
			"warning": {RGB: [3]int{235, 203, 139}, ANSI16: 3, ANSI256: 180},
			"error":   {RGB: [3]int{191, 97, 106}, ANSI16: 1, ANSI256: 131},

			// Content Relations
			"link":    {RGB: [3]int{136, 192, 208}, ANSI16: 6, ANSI256: 110},
			"visited": {RGB: [3]int{180, 142, 173}, ANSI16: 5, ANSI256: 139},

			// Boundaries
			"border":    {RGB: [3]int{76, 86, 106}, ANSI16: 8, ANSI256: 240},
			"focus":     {RGB: [3]int{136, 192, 208}, ANSI16: 6, ANSI256: 110},
			"select":    {RGB: [3]int{129, 161, 193}, ANSI16: 4, ANSI256: 109},
			"highlight": {RGB: [3]int{235, 203, 139}, ANSI16: 3, ANSI256: 180},

			// Disabled
			"disabled-bg": {RGB: [3]int{59, 66, 82}, ANSI16: 8, ANSI256: 237},
			"disabled-fg": {RGB: [3]int{97, 110, 136}, ANSI16: 8, ANSI256: 244},

			// System UI
			"scrollbar": {RGB: [3]int{76, 86, 106}, ANSI16: 8, ANSI256: 240},
			"shadow":    {RGB: [3]int{30, 32, 40}, ANSI16: 0, ANSI256: 232},
			"caret":     {RGB: [3]int{236, 239, 244}, ANSI16: 7, ANSI256: 253},
		},
	}
}

// draculaTheme implements the Dracula color scheme
func draculaTheme() ThemePreset {
	return ThemePreset{
		Name: "dracula",
		Colors: map[string]ColorValue{
			// Layer System
			"bg":      {RGB: [3]int{40, 42, 54}, ANSI16: 0, ANSI256: 236},
			"surface": {RGB: [3]int{52, 55, 70}, ANSI16: 8, ANSI256: 238},
			"overlay": {RGB: [3]int{36, 39, 51}, ANSI16: 0, ANSI256: 235},

			// Typography
			"text":        {RGB: [3]int{248, 248, 242}, ANSI16: 7, ANSI256: 255},
			"muted":       {RGB: [3]int{98, 114, 164}, ANSI16: 8, ANSI256: 103},
			"placeholder": {RGB: [3]int{98, 114, 164}, ANSI16: 8, ANSI256: 103},

			// Brand & Action
			"primary":   {RGB: [3]int{189, 147, 249}, ANSI16: 5, ANSI256: 141},
			"secondary": {RGB: [3]int{139, 233, 253}, ANSI16: 4, ANSI256: 111},
			"accent":    {RGB: [3]int{255, 121, 198}, ANSI16: 13, ANSI256: 212},

			// State
			"success": {RGB: [3]int{80, 250, 123}, ANSI16: 2, ANSI256: 84},
			"warning": {RGB: [3]int{241, 250, 140}, ANSI16: 3, ANSI256: 228},
			"error":   {RGB: [3]int{255, 85, 85}, ANSI16: 1, ANSI256: 203},

			// Content Relations
			"link":    {RGB: [3]int{139, 233, 253}, ANSI16: 4, ANSI256: 111},
			"visited": {RGB: [3]int{189, 147, 249}, ANSI16: 5, ANSI256: 141},

			// Boundaries
			"border":    {RGB: [3]int{68, 71, 90}, ANSI16: 8, ANSI256: 60},
			"focus":     {RGB: [3]int{189, 147, 249}, ANSI16: 5, ANSI256: 141},
			"select":    {RGB: [3]int{255, 121, 198}, ANSI16: 13, ANSI256: 212},
			"highlight": {RGB: [3]int{241, 250, 140}, ANSI16: 3, ANSI256: 228},

			// Disabled
			"disabled-bg": {RGB: [3]int{52, 55, 70}, ANSI16: 8, ANSI256: 238},
			"disabled-fg": {RGB: [3]int{98, 114, 164}, ANSI16: 8, ANSI256: 103},

			// System UI
			"scrollbar": {RGB: [3]int{68, 71, 90}, ANSI16: 8, ANSI256: 60},
			"shadow":    {RGB: [3]int{28, 29, 38}, ANSI16: 0, ANSI256: 234},
			"caret":     {RGB: [3]int{248, 248, 242}, ANSI16: 7, ANSI256: 255},
		},
	}
}

// gruvboxDarkTheme implements the Gruvbox dark color scheme
func gruvboxDarkTheme() ThemePreset {
	return ThemePreset{
		Name: "gruvbox-dark",
		Colors: map[string]ColorValue{
			// Layer System
			"bg":      {RGB: [3]int{40, 40, 40}, ANSI16: 0, ANSI256: 235},
			"surface": {RGB: [3]int{60, 56, 54}, ANSI16: 8, ANSI256: 237},
			"overlay": {RGB: [3]int{32, 30, 29}, ANSI16: 0, ANSI256: 234},

			// Typography
			"text":        {RGB: [3]int{235, 219, 178}, ANSI16: 7, ANSI256: 223},
			"muted":       {RGB: [3]int{146, 131, 116}, ANSI16: 8, ANSI256: 246},
			"placeholder": {RGB: [3]int{146, 131, 116}, ANSI16: 8, ANSI256: 246},

			// Brand & Action
			"primary":   {RGB: [3]int{131, 165, 152}, ANSI16: 4, ANSI256: 109},
			"secondary": {RGB: [3]int{142, 192, 124}, ANSI16: 6, ANSI256: 108},
			"accent":    {RGB: [3]int{211, 134, 155}, ANSI16: 5, ANSI256: 175},

			// State
			"success": {RGB: [3]int{184, 187, 38}, ANSI16: 2, ANSI256: 142},
			"warning": {RGB: [3]int{250, 189, 47}, ANSI16: 3, ANSI256: 214},
			"error":   {RGB: [3]int{251, 73, 52}, ANSI16: 1, ANSI256: 167},

			// Content Relations
			"link":    {RGB: [3]int{131, 165, 152}, ANSI16: 4, ANSI256: 109},
			"visited": {RGB: [3]int{211, 134, 155}, ANSI16: 5, ANSI256: 175},

			// Boundaries
			"border":    {RGB: [3]int{80, 73, 69}, ANSI16: 8, ANSI256: 239},
			"focus":     {RGB: [3]int{131, 165, 152}, ANSI16: 4, ANSI256: 109},
			"select":    {RGB: [3]int{211, 134, 155}, ANSI16: 5, ANSI256: 175},
			"highlight": {RGB: [3]int{250, 189, 47}, ANSI16: 3, ANSI256: 214},

			// Disabled
			"disabled-bg": {RGB: [3]int{60, 56, 54}, ANSI16: 8, ANSI256: 237},
			"disabled-fg": {RGB: [3]int{146, 131, 116}, ANSI16: 8, ANSI256: 246},

			// System UI
			"scrollbar": {RGB: [3]int{80, 73, 69}, ANSI16: 8, ANSI256: 239},
			"shadow":    {RGB: [3]int{28, 27, 26}, ANSI16: 0, ANSI256: 232},
			"caret":     {RGB: [3]int{235, 219, 178}, ANSI16: 7, ANSI256: 223},
		},
	}
}

// catppuccinMochaTheme implements the Catppuccin Mocha color scheme
func catppuccinMochaTheme() ThemePreset {
	return ThemePreset{
		Name: "catppuccin-mocha",
		Colors: map[string]ColorValue{
			// Layer System
			"bg":      {RGB: [3]int{30, 30, 46}, ANSI16: 0, ANSI256: 235},
			"surface": {RGB: [3]int{49, 50, 68}, ANSI16: 8, ANSI256: 238},
			"overlay": {RGB: [3]int{24, 24, 37}, ANSI16: 0, ANSI256: 234},

			// Typography
			"text":        {RGB: [3]int{205, 214, 244}, ANSI16: 7, ANSI256: 252},
			"muted":       {RGB: [3]int{108, 112, 134}, ANSI16: 8, ANSI256: 244},
			"placeholder": {RGB: [3]int{108, 112, 134}, ANSI16: 8, ANSI256: 244},

			// Brand & Action
			"primary":   {RGB: [3]int{137, 180, 250}, ANSI16: 4, ANSI256: 111},
			"secondary": {RGB: [3]int{203, 166, 247}, ANSI16: 5, ANSI256: 176},
			"accent":    {RGB: [3]int{245, 194, 231}, ANSI16: 13, ANSI256: 218},

			// State
			"success": {RGB: [3]int{166, 227, 161}, ANSI16: 2, ANSI256: 114},
			"warning": {RGB: [3]int{249, 226, 175}, ANSI16: 3, ANSI256: 223},
			"error":   {RGB: [3]int{243, 139, 168}, ANSI16: 1, ANSI256: 210},

			// Content Relations
			"link":    {RGB: [3]int{137, 180, 250}, ANSI16: 4, ANSI256: 111},
			"visited": {RGB: [3]int{203, 166, 247}, ANSI16: 5, ANSI256: 176},

			// Boundaries
			"border":    {RGB: [3]int{88, 91, 112}, ANSI16: 8, ANSI256: 239},
			"focus":     {RGB: [3]int{137, 180, 250}, ANSI16: 4, ANSI256: 111},
			"select":    {RGB: [3]int{245, 194, 231}, ANSI16: 13, ANSI256: 218},
			"highlight": {RGB: [3]int{249, 226, 175}, ANSI16: 3, ANSI256: 223},

			// Disabled
			"disabled-bg": {RGB: [3]int{49, 50, 68}, ANSI16: 8, ANSI256: 238},
			"disabled-fg": {RGB: [3]int{108, 112, 134}, ANSI16: 8, ANSI256: 244},

			// System UI
			"scrollbar": {RGB: [3]int{88, 91, 112}, ANSI16: 8, ANSI256: 239},
			"shadow":    {RGB: [3]int{22, 22, 33}, ANSI16: 0, ANSI256: 232},
			"caret":     {RGB: [3]int{205, 214, 244}, ANSI16: 7, ANSI256: 252},
		},
	}
}

// solarizedDarkTheme implements the Solarized Dark color scheme
func solarizedDarkTheme() ThemePreset {
	return ThemePreset{
		Name: "solarized-dark",
		Colors: map[string]ColorValue{
			// Layer System
			"bg":      {RGB: [3]int{0, 43, 54}, ANSI16: 0, ANSI256: 234},
			"surface": {RGB: [3]int{7, 54, 66}, ANSI16: 8, ANSI256: 237},
			"overlay": {RGB: [3]int{0, 36, 46}, ANSI16: 0, ANSI256: 233},

			// Typography
			"text":        {RGB: [3]int{238, 232, 213}, ANSI16: 7, ANSI256: 254},
			"muted":       {RGB: [3]int{131, 148, 150}, ANSI16: 8, ANSI256: 244},
			"placeholder": {RGB: [3]int{131, 148, 150}, ANSI16: 8, ANSI256: 244},

			// Brand & Action
			"primary":   {RGB: [3]int{38, 139, 210}, ANSI16: 4, ANSI256: 33},
			"secondary": {RGB: [3]int{42, 161, 152}, ANSI16: 6, ANSI256: 37},
			"accent":    {RGB: [3]int{108, 113, 196}, ANSI16: 5, ANSI256: 136},

			// State
			"success": {RGB: [3]int{133, 153, 0}, ANSI16: 2, ANSI256: 64},
			"warning": {RGB: [3]int{181, 137, 0}, ANSI16: 3, ANSI256: 136},
			"error":   {RGB: [3]int{220, 50, 47}, ANSI16: 1, ANSI256: 160},

			// Content Relations
			"link":    {RGB: [3]int{38, 139, 210}, ANSI16: 4, ANSI256: 33},
			"visited": {RGB: [3]int{108, 113, 196}, ANSI16: 5, ANSI256: 136},

			// Boundaries
			"border":    {RGB: [3]int{88, 110, 117}, ANSI16: 8, ANSI256: 240},
			"focus":     {RGB: [3]int{38, 139, 210}, ANSI16: 4, ANSI256: 33},
			"select":    {RGB: [3]int{42, 161, 152}, ANSI16: 6, ANSI256: 37},
			"highlight": {RGB: [3]int{181, 137, 0}, ANSI16: 3, ANSI256: 136},

			// Disabled
			"disabled-bg": {RGB: [3]int{7, 54, 66}, ANSI16: 8, ANSI256: 237},
			"disabled-fg": {RGB: [3]int{131, 148, 150}, ANSI16: 8, ANSI256: 244},

			// System UI
			"scrollbar": {RGB: [3]int{88, 110, 117}, ANSI16: 8, ANSI256: 240},
			"shadow":    {RGB: [3]int{0, 30, 38}, ANSI16: 0, ANSI256: 232},
			"caret":     {RGB: [3]int{238, 232, 213}, ANSI16: 7, ANSI256: 254},
		},
	}
}
