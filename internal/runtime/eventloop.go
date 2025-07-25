package runtime

import "sync"

// Global wait group to manage the event loop
var EventLoop = &sync.WaitGroup{}
