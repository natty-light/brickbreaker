package main

import (
	"github.com/engoengine/glm"
	"github.com/go-gl/gl/v2.1/gl"
)

type GameEntity struct {
	position   [2]float32
	vertices   [18]float32
	color      glm.Vec3
	vao        uint32
	vbo        uint32
	velocity   [2]float32
	flags      EntityFlags
	dimensions [2]float32
}

type EntityFlags struct {
	xVelScalar float32
	yVelScalar float32
	whoami     EntityType
}

type EntityType int

const (
	Paddle EntityType = 0
	Ball   EntityType = 1
	Brick  EntityType = 2
)

func (entity *GameEntity) GetTransformation() glm.Mat4 {
	return glm.Translate3D(entity.position[0], entity.position[1], 0.0)
}

func CreateGameEntity(position [2]float32, dimensions [2]float32, color glm.Vec3, vertices []float32, velocity [2]float32, whoami EntityType) *GameEntity {
	// Create openGL VAO and VBO, function found in main.go
	vao, vbo := CreateVAO(vertices)
	// Create entity
	entity := &GameEntity{position: position, dimensions: dimensions, color: color, vertices: [18]float32(vertices), vao: vao, vbo: vbo, velocity: velocity}
	// Create entities movement directives struct
	entity.flags = EntityFlags{0, 0, whoami}
	return entity
}

func CleanUpEntity(entity *GameEntity) {
	VAO, VBO := entity.vao, entity.vbo
	gl.DeleteVertexArrays(1, &VAO)
	gl.DeleteBuffers(1, &VBO)
}

func (entity *GameEntity) UpdatePosition(maxX, maxY int) {
	nextX, nextY := entity.position[0]+entity.flags.xVelScalar*entity.velocity[0], entity.position[1]+entity.flags.yVelScalar*entity.velocity[1]
	if nextX > (-1.0+entity.dimensions[0]/2.0) && nextX < (1.0-entity.dimensions[0]/2.0) {
		entity.position[0] = nextX
	}
	if nextY > (-1.0+entity.dimensions[1]/2.0) && nextY < (1.0-entity.dimensions[1]/2.0) {
		entity.position[1] = nextY
	}
}

func prepareSingleGameEntity(
	PosPosVertex []float32,
	PosNegVertex []float32,
	NegPosVertex []float32,
	NegNegVertex []float32,
	position [2]float32,
	velocity [2]float32,
	dimensions [2]float32,
	whoami EntityType,
) *GameEntity {
	var entityVertices []float32 = []float32{}
	entityVertices = append(entityVertices, PosPosVertex...)
	entityVertices = append(entityVertices, PosNegVertex...)
	entityVertices = append(entityVertices, NegNegVertex...)
	entityVertices = append(entityVertices, PosPosVertex...)
	entityVertices = append(entityVertices, NegPosVertex...)
	entityVertices = append(entityVertices, NegNegVertex...)

	entity := CreateGameEntity(position, dimensions, paddleColor, entityVertices, velocity, Brick)

	return entity
}

func drawEntity(entity *GameEntity, shaderProgram uint32) {
	entity.UpdatePosition(width, height)
	transformation := entity.GetTransformation()
	var objectColorLocation = gl.GetUniformLocation(shaderProgram, gl.Str("objectColor\x00"))
	var objectTransformationLocation = gl.GetUniformLocation(shaderProgram, gl.Str("transform\x00"))
	gl.Uniform3fv(objectColorLocation, 1, &brickColor[0])
	gl.UniformMatrix4fv(objectTransformationLocation, 1, false, &transformation[0])

	// perform rendering
	gl.UseProgram(shaderProgram)                                  // ensure the right shader program is being used
	gl.BindVertexArray(entity.vao)                                // bind data
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(entity.vertices)/3)) // perform draw call
	gl.BindVertexArray(0)                                         // unbind data (so we don't mistakenly use/modify it)
}
