package main

import (
	"github.com/engoengine/glm"
	"github.com/go-gl/gl/v2.1/gl"
)

type Brick struct {
	xPos     float32
	yPos     float32
	vertices [18]float32
	color    glm.Vec3
	vao      uint32
	vbo      uint32
}

func (brick *Brick) GetTransformation() glm.Mat4 {
	return glm.Translate3D(brick.xPos, brick.yPos, 0.0)
}

func CreateBrick(xPos float32, yPos float32, color glm.Vec3, vertices []float32) *Brick {
	vao, vbo := CreateVAO(vertices)
	brick := &Brick{xPos: xPos, yPos: yPos, color: color, vertices: [18]float32(vertices), vao: vao, vbo: vbo}
	return brick
}

func CleanUpBrick(brick *Brick) {
	VAO, VBO := brick.vao, brick.vbo
	gl.DeleteVertexArrays(1, &VAO)
	gl.DeleteBuffers(1, &VBO)
}
