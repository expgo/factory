package factory

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type animal struct {
	Name string
}

func (a *animal) Init() {
	a.Name = "animal"
}

type cat struct {
	animal // 嵌入 animal 结构体
	Name   string
}

func (c *cat) Init() {
	c.animal.Init()

	c.Name = "cat"
}

type dog struct {
	animal // 嵌入 animal 结构体
}

func (d *dog) Init() {
	d.animal.Init()

	d.Name = "dog"
}

func (c *cat) Meow() string {
	return "meow"
}

func TestSingletonBuilder(t *testing.T) {
	var c = Singleton[cat]().Getter()

	assert.Equal(t, "meow", c().Meow())
	assert.Equal(t, "cat", c().Name)
	assert.Equal(t, "animal", c().animal.Name)

	var d = Singleton[dog]().Getter()

	assert.Equal(t, "dog", d().Name)
	assert.Equal(t, "dog", d().animal.Name)
}
