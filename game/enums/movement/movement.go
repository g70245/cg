package movement

type Mode string

const (
	None                     Mode = "None"
	DIAGONAL                 Mode = "Diagonal"
	BIASED_DIAGONAL          Mode = "B. Diagonal"
	REVERSED_DIAGONAL        Mode = "Reversed Diagonal"
	BIASED_REVERSED_DIAGONAL Mode = "B. Reversed Diagonal"
	HYBRID_DIAGONAL          Mode = "Hybrid Diagonal"
)
