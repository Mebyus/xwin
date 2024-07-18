package x11

const (
	StatusFailed  = 0
	StatusSuccess = 1
	StatusAuth    = 2

	RequestCreateWindow = 1
	RequestMapWindow    = 8
	RequestImageText8   = 76
	RequestOpenFont     = 45
	RequestCreateGc     = 55

	EventFlagKeyPress   = 0x00000001
	EventFlagKeyRelease = 0x00000002
	EventFlagExposure   = 0x8000

	WindowClassCopyFromParent = 0
	WindowClassInputOutput    = 1
	WindowClassInputOnly      = 2

	FlagBackgroundPixel = 0x00000002
	FlagWinEvent        = 0x00000800
)
