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
}

func (entity *GameEntity) GetTransformation() glm.Mat4 {
	return glm.Translate3D(entity.xPos, entity.yPos, 0.0)
}

func CreateBrick(xPos float32, yPos float32, color glm.Vec3, vertices []float32) *GameEntity {
	vao, vbo := CreateVAO(vertices)
	entity := &GameEntity{xPos: xPos, yPos: yPos, color: color, vertices: [18]float32(vertices), vao: vao, vbo: vbo}
	return entity
}

func CleanUpEntity(entity *GameEntity) {
	VAO, VBO := entity.vao, entity.vbo
	gl.DeleteVertexArrays(1, &VAO)
	gl.DeleteBuffers(1, &VBO)
}
