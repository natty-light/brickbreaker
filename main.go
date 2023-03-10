package main

import (
	"fmt"
	"log"
	"runtime"
	"unsafe"

	"github.com/engoengine/glm"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

var (
	// global rotation
	width, height         int = 800, 800
	bricks                []*GameEntity
	brickDimensions       = [2]float32{.2, .1}
	paddle                *GameEntity
	paddleDimensions      = [2]float32{.3, .05}
	paddleInitialPosition = [2]float32{0, -.75}
	paddleVelocity        = [2]float32{.015, 0}
	ball                  *GameEntity
	ballInitialPosition   = [2]float32{0, -.55}
	ballDimensions        = [2]float32{.025, .025}
	ballInitialVelocity   = [2]float32{0.01, .01}
	brickColor            = glm.Vec3{1.0, 1.0, 1.0}
	paddleColor           = glm.Vec3{1.0, 1.0, 1.0}
	vertexShaderSource    = `
#version 410 core
layout (location = 0) in vec3 position;

uniform mat4 transform;

void main()
{
    gl_Position = transform * vec4(position.x, position.y, position.z, 1.0);
}
`
	fragmentShaderSource = `
#version 410 core

uniform vec3 objectColor;

out vec4 color;

void main()
{
	color = vec4(objectColor, 1.0);
}
`
	// Vertex definitions
	brickVertexPosPos = []float32{brickDimensions[0] / 2.0, brickDimensions[1] / 2.0, 0}
	brickVertexPosNeg = []float32{brickDimensions[0] / 2.0, -brickDimensions[1] / 2.0, 0}
	brickVertexNegNeg = []float32{-brickDimensions[0] / 2.0, -brickDimensions[1] / 2.0, 0}
	brickVertexNegPos = []float32{-brickDimensions[0] / 2.0, brickDimensions[1] / 2.0, 0}

	paddleVertexPosPos = []float32{paddleDimensions[0] / 2.0, paddleDimensions[1] / 2.0, 0}
	paddleVertexPosNeg = []float32{paddleDimensions[0] / 2.0, -paddleDimensions[1] / 2.0, 0}
	paddleVertexNegNeg = []float32{-paddleDimensions[0] / 2.0, -paddleDimensions[1] / 2.0, 0}
	paddleVertexNegPos = []float32{-paddleDimensions[0] / 2.0, paddleDimensions[1] / 2.0, 0}

	ballVertexPosPos = []float32{ballDimensions[0] / 2.0, ballDimensions[1] / 2.0, 0}
	ballVertexPosNeg = []float32{ballDimensions[0] / 2.0, -ballDimensions[1] / 2.0, 0}
	ballVertexNegNeg = []float32{-ballDimensions[0] / 2.0, -ballDimensions[1] / 2.0, 0}
	ballVertexNegPos = []float32{-ballDimensions[0] / 2.0, ballDimensions[1] / 2.0, 0}
)

type getGlParam func(uint32, uint32, *int32)
type getInfoLog func(uint32, int32, *int32, *uint8)

func checkGlError(glObject uint32, errorParam uint32, getParamFn getGlParam,
	getInfoLogFn getInfoLog, failMsg string) {

	var success int32
	getParamFn(glObject, errorParam, &success)
	if success != 1 {
		var infoLog [512]byte
		getInfoLogFn(glObject, 512, nil, (*uint8)(unsafe.Pointer(&infoLog)))
		log.Fatalln(failMsg, "\n", string(infoLog[:512]))
	}
}

func checkShaderCompileErrors(shader uint32) {
	checkGlError(shader, gl.COMPILE_STATUS, gl.GetShaderiv, gl.GetShaderInfoLog,
		"ERROR::SHADER::COMPILE_FAILURE")
}

func checkProgramLinkErrors(program uint32) {
	checkGlError(program, gl.LINK_STATUS, gl.GetProgramiv, gl.GetProgramInfoLog,
		"ERROR::PROGRAM::LINKING_FAILURE")
}

func compileShaders(vertShaderSource string, fragShaderSource string) []uint32 {
	// create the vertex shader
	vertexShader := gl.CreateShader(gl.VERTEX_SHADER)
	shaderSourceChars, freeVertexShaderFunc := gl.Strs(vertShaderSource)
	gl.ShaderSource(vertexShader, 1, shaderSourceChars, nil)
	gl.CompileShader(vertexShader)
	checkShaderCompileErrors(vertexShader)

	// create the fragment shader
	fragmentShader := gl.CreateShader(gl.FRAGMENT_SHADER)
	shaderSourceChars, freeFragmentShaderFunc := gl.Strs(fragShaderSource)
	gl.ShaderSource(fragmentShader, 1, shaderSourceChars, nil)
	gl.CompileShader(fragmentShader)
	checkShaderCompileErrors(fragmentShader)

	defer freeFragmentShaderFunc()
	defer freeVertexShaderFunc()

	return []uint32{vertexShader, fragmentShader}
}

/*
 * Link the provided shaders in the order they were given and return the linked program.
 */
func linkShaders(shaders []uint32) uint32 {
	program := gl.CreateProgram()
	for _, shader := range shaders {
		gl.AttachShader(program, shader)
	}
	gl.LinkProgram(program)
	checkProgramLinkErrors(program)

	// shader objects are not needed after they are linked into a program object
	for _, shader := range shaders {
		gl.DeleteShader(shader)
	}

	return program
}

func CreateVAO(vertices []float32) (VAO uint32, VBO uint32) {
	gl.GenVertexArrays(1, &VAO)
	gl.GenBuffers(1, &VBO)

	// Bind the Vertex Array Object first, then bind and set vertex buffer(s) and attribute pointers()
	gl.BindVertexArray(VAO)

	// copy vertices data into VBO (it needs to be bound first)
	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// specify the format of our vertex input
	// (shader) input 0
	// vertex has size 3
	// vertex items are of type FLOAT
	// do not normalize (already done)
	// stride of 3 * sizeof(float) (separation of vertices)
	// offset of where the position data starts (0 for the beginning)
	gl.VertexAttribPointerWithOffset(0, 3, gl.FLOAT, false, 3*4, 0)
	gl.EnableVertexAttribArray(0)

	// unbind the VAO (safe practice so we don't accidentally (mis)configure it later)
	gl.BindVertexArray(0)

	return VAO, VBO
}

func init() {
	runtime.LockOSThread()
}

func reshape(window *glfw.Window, w, h int) {
	gl.ClearColor(1, 1, 1, 1)
	/* Establish viewing area to cover entire window. */
	gl.Viewport(0, 0, int32(w), int32(h))
	/* PROJECTION Matrix mode. */
	gl.MatrixMode(gl.PROJECTION)
	/* Reset project matrix. */
	gl.LoadIdentity()
	/* Map abstract coords directly to window coords. */
	gl.Ortho(0, float64(w), 0, float64(h), -1, 1)
	/* Invert Y axis so increasing Y goes down. */
	gl.Scalef(1, -1, 1)
	/* Shift origin up to upper-left corner. */
	gl.Translatef(0, float32(-h), 0)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.Disable(gl.DEPTH_TEST)
	width, height = w, h
}

func main() {
	err := glfw.Init()
	if err != nil {
		panic(err)
	}
	defer glfw.Terminate()
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	window, err := glfw.CreateWindow(width, height, "Brick", nil, nil)
	if err != nil {
		panic(err)
	}

	window.MakeContextCurrent()
	window.SetSizeCallback(reshape)
	window.SetKeyCallback(onKey)
	window.SetCharCallback(onChar)

	glfw.SwapInterval(-1)
	err = gl.Init()
	if err != nil {
		panic(err)
	}

	reshape(window, width, height)
	// Prepare Game Objects
	bricks = prepareBricks()
	paddle = prepareSingleGameEntity(
		paddleVertexPosPos, paddleVertexPosNeg,
		paddleVertexNegPos, paddleVertexNegNeg,
		paddleInitialPosition,
		paddleVelocity,
		paddleDimensions,
		0)
	ball = prepareSingleGameEntity(
		ballVertexPosPos, ballVertexPosNeg,
		ballVertexNegPos, ballVertexNegNeg,
		ballInitialPosition,
		ballInitialVelocity,
		ballDimensions,
		1)
	ball.flags.xVelScalar, ball.flags.yVelScalar = 1, 1
	fmt.Print(ball.flags.whoami)

	// Compile shaders
	shaders := compileShaders(vertexShaderSource, fragmentShaderSource)
	shaderProgram := linkShaders(shaders)

	for !window.ShouldClose() {
		gl.ClearColor(0.0, 0.0, 0.0, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		// Call draw function on game entities
		for _, brick := range bricks {
			if brick.flags.enabled {
				drawEntity(brick, shaderProgram)
				checkEntityCollision(brick, ball, brickCollisionCallback)
			}
		}
		checkEntityCollision(paddle, ball, func(ge *GameEntity) {})
		paddle.UpdatePosition(width, height)
		ball.UpdatePosition(width, height)
		drawEntity(paddle, shaderProgram)
		drawEntity(ball, shaderProgram)
		// end of draw loop
		// swap in the rendered buffer
		window.SwapBuffers()
		glfw.PollEvents()
	}
	// Clean Up
	for _, brick := range bricks {
		CleanUpEntity(brick)
	}
	CleanUpEntity(paddle)
	CleanUpEntity(ball)
}

func onChar(w *glfw.Window, char rune) {
	log.Println(char)
}

// Keyboard key callback
func onKey(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	switch {
	case key == glfw.KeyEscape && action == glfw.Press,
		key == glfw.KeyQ && action == glfw.Press:
		w.SetShouldClose(true)
	case key == glfw.KeyLeft && action == glfw.Press:
		paddle.flags.xVelScalar += -1.0
	case key == glfw.KeyRight && action == glfw.Press:
		paddle.flags.xVelScalar += 1.0
	case key == glfw.KeyRight && action == glfw.Release:
		paddle.flags.xVelScalar -= 1.0
	case key == glfw.KeyLeft && action == glfw.Release:
		paddle.flags.xVelScalar -= -1.0
	}
}

func prepareBricks() []*GameEntity {
	var brickVertices []float32 = []float32{}
	brickVertices = append(brickVertices, brickVertexPosPos...)
	brickVertices = append(brickVertices, brickVertexPosNeg...)
	brickVertices = append(brickVertices, brickVertexNegNeg...)
	brickVertices = append(brickVertices, brickVertexPosPos...)
	brickVertices = append(brickVertices, brickVertexNegPos...)
	brickVertices = append(brickVertices, brickVertexNegNeg...)

	bricks := []*GameEntity{}
	for j := 0; j < 4; j++ {
		for i := 0; i < 7; i++ {
			var x, y float32 = -0.75 + float32(i)*.25, .8 - 0.15*float32(j)
			var brick *GameEntity = CreateGameEntity([2]float32{x, y}, brickDimensions, brickColor, brickVertices, [2]float32{0.0, 0.0}, 2, true)
			bricks = append(bricks, brick)
		}
	}
	return bricks
}

func brickCollisionCallback(entity *GameEntity) {
	entity.flags.enabled = false
}
