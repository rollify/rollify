package ui

import "github.com/rollify/rollify/internal/model"

var diceQuantity = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

type die struct {
	Color string
	SVG   string
	model.DieType
	RangedQuantity []int // We need a 0 value slice to range over quantity.
}

const (
	dieSVGD4  = `M90.68,68.81,46.28,6.93a2,2,0,0,0-3.4.25L13.16,64.86a2,2,0,0,0,.17,2.1L32.56,93.18a2,2,0,0,0,1.61.82,2.11,2.11,0,0,0,.74-.14l54.89-22a2,2,0,0,0,.88-3ZM17.29,65.59,42.61,16.44l-8.94,71.5ZM35.51,89.31l9.64-77.1L86,69.07Zm30.81-28-2.47.82,2.24,6.41-4.53,1.66L59.51,63.6,50,66.77,48.82,61.6,53,44.51l4.35-1,5,14.45,2.41-.74Zm-8.15-2.08-3.3-10.67L52.44,61Z`
	dieSVGD6  = `M86.08,30.84,49.32,7.32A2,2,0,0,0,47,7.4l-33.24,25A2,2,0,0,0,13,34.1l1,30a2,2,0,0,0,.83,1.55L51.6,92.09a2,2,0,0,0,1.17.38A2,2,0,0,0,54.12,92L85.35,63.48A2,2,0,0,0,86,62.07l1-29.48A2,2,0,0,0,86.08,30.84ZM48.34,11.43l34.21,21.9-29.85,26L17.56,34.61ZM18,63l-.88-26.23L51.77,61.08V87.29Zm35.8,23.9V61L82.89,35.68,82,61.09ZM47.46,28.31a2.69,2.69,0,0,0-1.28-.5,2.46,2.46,0,0,0-2.29.87c-.94,1.05-.94,2.36,0,3.94a17.91,17.91,0,0,0,3,3.29,4.79,4.79,0,0,1,0-2.47,5.89,5.89,0,0,1,1.23-2.19A6.33,6.33,0,0,1,53,29a9.14,9.14,0,0,1,5.75,2.12,10.92,10.92,0,0,1,3.85,5.33c.71,2.14.24,4.28-1.43,6.42a6.79,6.79,0,0,1-7.39,2.47,18.48,18.48,0,0,1-7.55-4.12A42,42,0,0,1,42.7,38a14.83,14.83,0,0,1-3.06-4.57,7.54,7.54,0,0,1-.49-3.69,6.07,6.07,0,0,1,1.65-3.39,7.17,7.17,0,0,1,4.56-2.5,7.28,7.28,0,0,1,4.89,1.3Zm7.76,13a3,3,0,0,0,2.65-1.14,2.67,2.67,0,0,0,.47-2.65,5.76,5.76,0,0,0-2.06-2.71A5.4,5.4,0,0,0,53,33.39a3,3,0,0,0-2.44,1.07,2.92,2.92,0,0,0-.71,2,4.7,4.7,0,0,0,2,3.39A5.57,5.57,0,0,0,55.22,41.28Z`
	dieSVGD8  = `M46.37,35.69a4.13,4.13,0,0,1,.54-3.21,8,8,0,0,1-3.75-.81A9,9,0,0,1,40.91,30a5,5,0,0,1-1.49-4.4q.38-2.39,3.48-4.16A9.32,9.32,0,0,1,49,20.11a8.09,8.09,0,0,1,4.89,2.46,7.07,7.07,0,0,1,1.36,2.19A3.54,3.54,0,0,1,55,27.67a9.41,9.41,0,0,1,4.18.71,9.55,9.55,0,0,1,3.39,2.37,5.93,5.93,0,0,1,1.76,5.06q-.42,2.74-3.88,4.72A9.64,9.64,0,0,1,48.2,39,7.12,7.12,0,0,1,46.37,35.69ZM55,38a4,4,0,0,0,2.88-.5,2.56,2.56,0,0,0,1.49-2,3.5,3.5,0,0,0-1.17-2.56,5.21,5.21,0,0,0-2.8-1.69,4.09,4.09,0,0,0-2.86.52,2.61,2.61,0,0,0-1.49,2,3.49,3.49,0,0,0,1.19,2.6A5.16,5.16,0,0,0,55,38Zm-7.56-8.54A3.6,3.6,0,0,0,49.91,29a2.28,2.28,0,0,0,1.33-1.74,2.63,2.63,0,0,0-.85-2,4.13,4.13,0,0,0-2.27-1.33,3.66,3.66,0,0,0-2.5.48,2.36,2.36,0,0,0-1.34,1.72,2.69,2.69,0,0,0,.9,2.11A3.94,3.94,0,0,0,47.39,29.42ZM93,26.72a2,2,0,0,0-1.19-1L27.64,6.69A2,2,0,0,0,25.15,8L6.09,70.24A2,2,0,0,0,7.4,72.73L71.62,92.8a1.8,1.8,0,0,0,.6.1,1.92,1.92,0,0,0,.93-.24,2,2,0,0,0,1-1.19l19-63.22A2,2,0,0,0,93,26.72Zm-5.41,2L37.8,58.2,28.59,11.15ZM27.26,14.82l8.68,44.33-25.2,9.57ZM12.57,70.16l24.31-9.23,29.35,26ZM70.88,88.38,38.78,59.94,88.26,30.59Z`
	dieSVGD10 = `M43.35,35l-.9-2.36a19.88,19.88,0,0,0,2.45-1.14,3.65,3.65,0,0,0,1.5-1.55,2.83,2.83,0,0,0,.22-1.45,2.64,2.64,0,0,0-.14-.7l3.09-1.1,8.75,18.58L53.91,47,48,33.27ZM63.68,24a16.91,16.91,0,0,1,5.37,5.72c1.86,2.93,2.84,5.49,2.86,7.61s-1.31,3.84-4,4.88a7.27,7.27,0,0,1-7.19-.61,17.76,17.76,0,0,1-5.28-6.84c-1.44-3-2-5.32-1.77-7.15s1.53-3.07,3.87-3.91A7,7,0,0,1,63.68,24Zm-1,14a3,3,0,0,0,3.4.9,2,2,0,0,0,1.34-2.68,19.72,19.72,0,0,0-2.3-5.06,19.23,19.23,0,0,0-3.19-4.33,2.73,2.73,0,0,0-3-.62,1.89,1.89,0,0,0-1.42,2.21,16.38,16.38,0,0,0,1.91,4.83A20.76,20.76,0,0,0,62.65,38ZM94.71,40l-6-10a1.93,1.93,0,0,0-.93-.81l-52-22a2,2,0,0,0-2.56.92l-27,52a2,2,0,0,0,.14,2.07l7,10a2,2,0,0,0,.89.7l52,21a2,2,0,0,0,2.53-.94l26-51A2,2,0,0,0,94.71,40Zm-54.8,7.14-3.63-35.4L84.79,32.24,70.55,50.94ZM10.33,60.85l24.2-46.62L38,47.61,19.78,64.83l-9.41-3.92Zm6,8.52L12.51,64l7,2.92h0l25.37,14ZM66.3,89l-.25.49-5.82-2.35L21.69,65.79,39.34,49.05,69.9,52.87Zm2.5-4.91L72,52.38,86.25,33.64l4.46,7.44Z`
	dieSVGD12 = `M34.89,32.13l-2-2.19a21.09,21.09,0,0,0,2.51-1.85,3.42,3.42,0,0,0,1.18-2,2.54,2.54,0,0,0-.29-1.52,2.67,2.67,0,0,0-.43-.67L39,22,55.92,38.11l-4.17,3.06-12-12.12Zm20.33.31c-.68-1.49-.7-3.49-.07-6a21.7,21.7,0,0,0,.76-4.13,3.38,3.38,0,0,0-1.34-2.66,4.63,4.63,0,0,0-2.32-1,3.64,3.64,0,0,0-2.43.53c-1.07.65-1.45,1.4-1.13,2.25a6.11,6.11,0,0,0,1.58,2L46.69,25.7a8.49,8.49,0,0,1-2.38-3.52c-.54-1.95.4-3.62,2.79-5a10.07,10.07,0,0,1,5.69-1.5,9.19,9.19,0,0,1,5.33,2A5.61,5.61,0,0,1,60.37,21a8.86,8.86,0,0,1-.22,3.47l-.38,1.79a21.92,21.92,0,0,0-.44,2.44,3.57,3.57,0,0,0,.17,1.47l7.77-5.48,3.4,2.63-12.6,9.25A11.58,11.58,0,0,1,55.22,32.44ZM92.85,47.33l-8.59-21a2,2,0,0,0-.65-.85L60.72,8.33A2,2,0,0,0,59.22,8L27.73,12.72a2,2,0,0,0-1.18.64L17,23.85a2.15,2.15,0,0,0-.38.61L6.14,51.18a2,2,0,0,0,0,1.36l3.34,10a2.18,2.18,0,0,0,.4.69l21,23.85a2,2,0,0,0,.47.39l7.15,4.29a2,2,0,0,0,1,.29,2.31,2.31,0,0,0,.37,0l25.76-4.77a2.42,2.42,0,0,0,.57-.2L82.4,78.48a2.06,2.06,0,0,0,1-1.13L92.9,48.72A2,2,0,0,0,92.85,47.33ZM80,27.79,56.47,47,25.39,35.07l4.26-18.59L59,12ZM20.22,26.27l6.87-7.55L23.26,35.37,12.14,58l-2-6.06ZM13.11,60.91,13,60.72,24.71,37,55.76,48.82,60.3,76.94,33.8,84.33l-.18-.11ZM64.58,83.39,39.86,88l-3.7-2.22,25.45-7.09,14.6-1.42ZM80,74.86l-17.7,1.77L57.73,48.5,81.19,29.39l7.68,18.77Z`
	dieSVGD20 = `M91.9,61.37l-14-42a2,2,0,0,0-1.33-1.29l-37-11a2,2,0,0,0-2,.51l-31,31a2,2,0,0,0-.48,2.06l15,44A2,2,0,0,0,22.73,86l36,5L59,91a2,2,0,0,0,1.31-.49l31-27A2,2,0,0,0,91.9,61.37ZM24.44,82,31.6,47.05l33.08,34ZM77.26,31.69,32.33,43.61l7.06-31.75Zm-47,12.12L11.16,39.66,36.88,14ZM66.38,79.92,32.91,45.53,78.38,33.46l.12.35-12,46ZM77,29.29l-30.52-16,28,8.31ZM10.66,41.6l19.16,4.17L23.16,78.26ZM58.37,86.89,35.5,83.72l27.23-.62Zm10.85-9.44L79.66,37.29l8,24.08ZM41.64,35.5a7,7,0,0,1,1.85-3.82,16.93,16.93,0,0,0,1.84-2.58A2.44,2.44,0,0,0,45.4,27a2.67,2.67,0,0,0-1.13-1.28,2.34,2.34,0,0,0-1.74-.21c-.9.2-1.38.64-1.46,1.31a4.35,4.35,0,0,0,.36,1.74l-3,.7a6,6,0,0,1-.42-3c.26-1.51,1.4-2.47,3.4-2.91a6.62,6.62,0,0,1,4.21.31,5.27,5.27,0,0,1,2.73,2.71,3.82,3.82,0,0,1,.21,2.86,7.46,7.46,0,0,1-1.35,2.28l-.85,1.08q-.81,1-1.08,1.47A2.81,2.81,0,0,0,45,35l6.62-1.74,1.11,2.45L42.09,38.6A6.75,6.75,0,0,1,41.64,35.5ZM56.57,21.81A13,13,0,0,1,60,26.36a10.48,10.48,0,0,1,1.47,5.28c-.1,1.4-1.06,2.36-2.87,2.85a5.11,5.11,0,0,1-4.76-.82,13,13,0,0,1-3.53-4.93,9.78,9.78,0,0,1-1.2-5.25c.2-1.3,1.19-2.14,3-2.53A5.18,5.18,0,0,1,56.57,21.81Zm-1.26,9.66a2.13,2.13,0,0,0,2.25.81,1.29,1.29,0,0,0,1-1.66,13.72,13.72,0,0,0-1.33-3.57,15,15,0,0,0-2.06-3.34A2.09,2.09,0,0,0,53.07,23,1.24,1.24,0,0,0,52,24.46,12.67,12.67,0,0,0,53.23,28,15.33,15.33,0,0,0,55.31,31.47Z`
)

var (
	dieD4  = die{DieType: model.DieTypeD4, Color: "currentColor", RangedQuantity: diceQuantity, SVG: dieSVGD4}
	dieD6  = die{DieType: model.DieTypeD6, Color: "currentColor", RangedQuantity: diceQuantity, SVG: dieSVGD6}
	dieD8  = die{DieType: model.DieTypeD8, Color: "currentColor", RangedQuantity: diceQuantity, SVG: dieSVGD8}
	dieD10 = die{DieType: model.DieTypeD10, Color: "currentColor", RangedQuantity: diceQuantity, SVG: dieSVGD10}
	dieD12 = die{DieType: model.DieTypeD12, Color: "currentColor", RangedQuantity: diceQuantity, SVG: dieSVGD12}
	dieD20 = die{DieType: model.DieTypeD20, Color: "currentColor", RangedQuantity: diceQuantity, SVG: dieSVGD20}
)