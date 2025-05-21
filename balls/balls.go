package balls

import (
	"math/rand"
	"unsafe"
)

type Velocity struct {
	X float32
	Y float32
}
type Circle struct {
	uid    uint8
	X      uint16
	Y      uint16
	Radius float32
	r ,g, b, a uint8
}

func (c *Circle) GetColor() (uint8, uint8, uint8, uint8) {
	return c.r, c.g, c.b, c.a
}

type Context struct {
	Width  int32
	Height int32

	Circles    []Circle
	Velocities []Velocity
}

func MakeContext(width, height int32) Context {
	c := Context{
		Width:  width,
		Height: height,
	}

	c.Circles = make([]Circle, 0)
	c.Velocities = make([]Velocity, 0)

	return c
}

func (c *Context) InitCircles(numCircles int) {
	for i := 0; i < numCircles; i++ {
		radius := float32(rand.Intn(50) + 10)
		x := uint16(rand.Intn(int(c.Width)-int(2*radius)) + int(radius))
		y := uint16(rand.Intn(int(c.Height)-int(2*radius)) + int(radius))

		velocity := Velocity{
		X: float32(rand.Intn(2)+1) * float32(1-2*rand.Intn(2)), // random value between -5 and 5, excluding 0
		Y: float32(rand.Intn(2)+1) * float32(1-2*rand.Intn(2)),
	}

		c.AddCircle(x, y, radius, velocity)
	}
}

func (c *Context) AddCircle(x, y uint16, radius float32, v Velocity) {
	r := uint8(rand.Intn(256))
	g := uint8(rand.Intn(256))
	b := uint8(rand.Intn(256))
	a := uint8(255)

	circle := Circle{
		X:      x,
		Y:      y,
		Radius: radius,
		r:      r,
		g:      g,
		b:      b,
		a:      a,
	}
	c.Circles = append(c.Circles, circle)

	c.Velocities = append(c.Velocities, v)
}

func (c *Context) UpdateCircles() []Circle{
	for i := 0; i < len(c.Circles); i++ {
		circle := &c.Circles[i]
		velocity := &c.Velocities[i]

		circle.X += uint16(velocity.X)
		circle.Y += uint16(velocity.Y)

		if float32(circle.X)-circle.Radius < 0 || float32(circle.X)+circle.Radius > float32(c.Width) {
			velocity.X = -velocity.X
		}
		if float32(circle.Y)-circle.Radius < 0 || float32(circle.Y)+circle.Radius > float32(c.Height) {
			velocity.Y = -velocity.Y
		}

	}

	return c.Circles
}

func (c *Context) ExportState() []byte {
	state := make([]byte, 0, len(c.Circles)*13)
	for _, circle := range c.Circles {
		// 1 byte uid, 2 bytes X, 2 bytes Y, 4 bytes Radius (float32), 4 bytes color (r,g,b,a)
		state = append(state, byte(circle.uid))
		state = append(state, byte(circle.X>>8), byte(circle.X))
		state = append(state, byte(circle.Y>>8), byte(circle.Y))
		radiusBytes := make([]byte, 4)
		r := circle.Radius
		bits := *(*uint32)(unsafe.Pointer(&r))
		radiusBytes[0] = byte(bits >> 24)
		radiusBytes[1] = byte(bits >> 16)
		radiusBytes[2] = byte(bits >> 8)
		radiusBytes[3] = byte(bits)
		state = append(state, radiusBytes...)
		// Add color (r, g, b, a)
		state = append(state, circle.r, circle.g, circle.b, circle.a)
	}
	return state
}

func ImportState(state []byte) []Circle {
	circles := make([]Circle, 0, len(state)/13)

	for i := 0; i < len(state); i += 13 {
		if i+13 > len(state) {
			break
		}
		circle := Circle{
			uid:    state[i],
			X:      uint16(state[i+1])<<8 | uint16(state[i+2]),
			Y:      uint16(state[i+3])<<8 | uint16(state[i+4]),
			Radius: float32FromBytes(state[i+5 : i+9]),
			r:      state[i+9],
			g:      state[i+10],
			b:      state[i+11],
			a:      state[i+12],
		}
		circles = append(circles, circle)

	}

	return circles
}
func uint32FromFloat32(f float32) uint32 {
	bits := *(*uint32)(unsafe.Pointer(&f))
	return bits
}

func float32FromBytes(b []byte) float32 {
	if len(b) != 4 {
		panic("byte slice must be 4 bytes long")
	}
	bits := uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
	return *(*float32)(unsafe.Pointer(&bits))
}

