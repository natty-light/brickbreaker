package main

import (
	"log"
	"runtime"
	"time"
	"unsafe"

	"github.com/engoengine/glm"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

var (
	// global rotation
	width, height      int = 800, 800
	objectColor            = glm.Vec3{1.0, 1.0, 1.0}
	vertexShaderSource     = `
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
	L1 = []float32{0.25, 0.25, 0}
	L2 = []float32{0.25, -0.25, 0}
	L3 = []float32{-0.25, -0.25, 0}
	L4 = []float32{-0.25, 0.25, 0}
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

	var vertices []float32 = []float32{}
	vertices = append(vertices, L1...)
	vertices = append(vertices, L2...)
	vertices = append(vertices, L3...)
	vertices = append(vertices, L1...)
	vertices = append(vertices, L4...)
	vertices = append(vertices, L3...)

	brick := CreateBrick(0.5, 0.5, objectColor, vertices)
	shaders := compileShaders(vertexShaderSource, fragmentShaderSource)
	shaderProgram := linkShaders(shaders)

	for !window.ShouldClose() {
		transformation := brick.GetTransformation()
		var objectColorLocation = gl.GetUniformLocation(shaderProgram, gl.Str("objectColor\x00"))
		var objectTransformationLocation = gl.GetUniformLocation(shaderProgram, gl.Str("transform\x00"))
		gl.Uniform3fv(objectColorLocation, 1, &objectColor[0])
		gl.UniformMatrix4fv(objectTransformationLocation, 1, false, &transformation[0])

		// perform rendering
		gl.ClearColor(0.0, 0.0, 0.0, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)
		gl.UseProgram(shaderProgram)                                 // ensure the right shader program is being used
		gl.BindVertexArray(brick.vao)                                // bind data
		gl.DrawArrays(gl.TRIANGLES, 0, int32(len(brick.vertices)/3)) // perform draw call
		gl.BindVertexArray(0)                                        // unbind data (so we don't mistakenly use/modify it)
		// end of draw loop

		// swap in the rendered buffer
		window.SwapBuffers()
		glfw.PollEvents()
		time.Sleep(16 * time.Millisecond)
	}
	CleanUpBrick(brick)
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
	}
}
