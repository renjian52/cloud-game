// Package savestates takes care of serializing and unserializing the game RAM
// to the host filesystem.
package nanoarch



func (na *naEmulator) GetLock() {
	//atomic.CompareAndSwapInt32(&na.saveLock, 0, 1)
	na.lock.Lock()
}

func (na *naEmulator) ReleaseLock() {
	//atomic.CompareAndSwapInt32(&na.saveLock, 1, 0)
	na.lock.Unlock()
}

// Save the current state to the filesystem. name is the name of the
// savestate file to save to, without extension.
func (na *naEmulator) Save() error {


	return nil
}

// Load the state from the filesystem
func (na *naEmulator) Load() error {
	return nil
}
