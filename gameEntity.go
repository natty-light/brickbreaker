package main

import (
	"github.com/engoengine/glm"
	"github.com/go-gl/gl/v2.1/gl"
)

type GameEntity struct {
	xPos     float32
	yPos     float32
	vertices [18]float32
	color    glm.Vec3
	vao      uint32
	vbo      uint32
	velocity [2]float32
	flags    EntityFlags
}

type EntityFlags struct {
	xVelScalar float32
	yVelScalar float32
}

func (entity *GameEntity) GetTransformation() glm.Mat4 {
	return glm.Translate3D(entity.xPos, entity.yPos, 0.0)
}

func CreateGameEntity(xPos float32, yPos float32, color glm.Vec3, vertices []float32, velocity [2]float32) *GameEntity {
	// Create openGL VAO and VBO, function found in main.go
	vao, vbo := CreateVAO(vertices)
	// Create entity
	entity := &GameEntity{xPos: xPos, yPos: yPos, color: color, vertices: [18]float32(vertices), vao: vao, vbo: vbo, velocity: velocity}
	// Create entities movement directives struct
	entity.flags = EntityFlags{0, 0}
	return entity
}

func CleanUpEntity(entity *GameEntity) {
	VAO, VBO := entity.vao, entity.vbo
	gl.DeleteVertexArrays(1, &VAO)
	gl.DeleteBuffers(1, &VBO)
}

func (entity *GameEntity) UpdatePosition(maxX, maxY int) {
	entity.xPos += entity.flags.xVelScalar * entity.velocity[0]
	entity.yPos += entity.flags.yVelScalar * entity.velocity[1]
}
