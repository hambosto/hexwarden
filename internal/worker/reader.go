package worker

import (
	"encoding/binary"
	"fmt"
	"io"
	"sync/atomic"
)

func (w *Worker) readForEncryption(reader io.Reader, tasks chan<- Task) error {
	var index uint32
	buffer := make([]byte, DefaultChunkSize)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read failed: %w", err)
		}

		data := make([]byte, n)
		copy(data, buffer[:n])

		tasks <- Task{Data: data, Index: atomic.LoadUint32(&index)}
		atomic.AddUint32(&index, 1)
	}

	return nil
}

func (w *Worker) readForDecryption(reader io.Reader, tasks chan<- Task) error {
	var index uint32
	var sizeBuffer [4]byte

	for {
		_, err := io.ReadFull(reader, sizeBuffer[:])
		if err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("chunk size read failed: %w", err)
		}

		chunkLen := binary.BigEndian.Uint32(sizeBuffer[:])

		data := make([]byte, chunkLen)
		if _, err := io.ReadFull(reader, data); err != nil {
			return fmt.Errorf("chunk data read failed: %w", err)
		}

		tasks <- Task{Data: data, Index: atomic.LoadUint32(&index)}
		atomic.AddUint32(&index, 1)
	}

	return nil
}
