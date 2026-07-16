package container

import (
	"encoding/json"
	"fmt"
	"io"

	"cg/game/battle"
)

func loadActionConfiguration(reader io.ReadCloser) (actionState battle.ActionState, err error) {
	defer closeActionConfiguration(reader, &err)

	data, err := io.ReadAll(reader)
	if err != nil {
		return actionState, fmt.Errorf("read file: %w", err)
	}
	if err := json.Unmarshal(data, &actionState); err != nil {
		return actionState, fmt.Errorf("invalid file format: %w", err)
	}
	return actionState, nil
}

func saveActionConfiguration(writer io.WriteCloser, actionState battle.ActionState) (err error) {
	defer closeActionConfiguration(writer, &err)

	data, err := json.Marshal(actionState)
	if err != nil {
		return fmt.Errorf("encode file: %w", err)
	}

	written, err := writer.Write(data)
	if err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	if written != len(data) {
		return fmt.Errorf("write file: %w", io.ErrShortWrite)
	}
	return nil
}

func closeActionConfiguration(closer io.Closer, operationErr *error) {
	if err := closer.Close(); err != nil && *operationErr == nil {
		*operationErr = fmt.Errorf("close file: %w", err)
	}
}
