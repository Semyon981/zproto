package zmux

// Конфиг
type Config struct {
	// Размер буфера для отправки в сокет.
	// Если значение меньше или равно 0, оно заменится дефолтным
	SendBufferSize int
	// Размер буферов, читающих из сокета.
	// У каждого канала он свой.
	// Если значение меньше или равно 0, оно заменится дефолтным
	RecvBuffersSize int

	// Максимальный размер фрейма, на который будут делиться данные
	// Если значение меньше или равно 0, оно заменится дефолтным
	FrameSize uint32
}

func (c *Config) normalize() {
	if c.RecvBuffersSize <= 0 {
		c.RecvBuffersSize = DefaultConfig.RecvBuffersSize
	}
	if c.SendBufferSize <= 0 {
		c.SendBufferSize = DefaultConfig.SendBufferSize
	}
	if c.FrameSize <= 0 {
		c.FrameSize = DefaultConfig.FrameSize
	}
}

var DefaultConfig = Config{
	SendBufferSize:  1024 * 16,
	RecvBuffersSize: 1024 * 2,
	FrameSize:       1024,
}
