package worker

import (
	"encoding/binary"
	"fmt"
	"io"
)

func (w *Worker) readTasks(input io.Reader) error {
	if w.config.Processing == Encryption {
		return w.readForEncryption(input)
	}
	return w.readForDecryption(input)
}

func (w *Worker) readForEncryption(reader io.Reader) error {
	buffer := make([]byte, w.config.ChunkSize)
	var index uint64

	for {
		select {
		case <-w.ctx.Done():
			return w.ctx.Err()
		default:
		}

		n, err := reader.Read(buffer)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read failed: %w", err)
		}

		// Create a copy of the data
		data := make([]byte, n)
		copy(data, buffer[:n])

		task := Task{
			Data:  data,
			Index: index,
		}

		select {
		case w.taskChan <- task:
			index++
		case <-w.ctx.Done():
			return w.ctx.Err()
		}
	}
}

func (w *Worker) readForDecryption(reader io.Reader) error {
	var sizeBuffer [4]byte
	var index uint64

	for {
		select {
		case <-w.ctx.Done():
			return w.ctx.Err()
		default:
		}

		_, err := io.ReadFull(reader, sizeBuffer[:])
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("chunk size read failed: %w", err)
		}

		chunkLen := binary.BigEndian.Uint32(sizeBuffer[:])
		if chunkLen == 0 {
			continue
		}

		data := make([]byte, chunkLen)
		if _, err := io.ReadFull(reader, data); err != nil {
			return fmt.Errorf("chunk data read failed: %w", err)
		}

		task := Task{
			Data:  data,
			Index: index,
		}

		select {
		case w.taskChan <- task:
			index++
		case <-w.ctx.Done():
			return w.ctx.Err()
		}
	}
}
