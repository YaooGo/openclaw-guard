// Scanner service - simple wrapper for API calls
// Actual scanning logic is in Electron main process

class ScannerService {
  async scan() {
    // This is handled by api.js
    throw new Error('Use api.scan() instead')
  }
}

export default new ScannerService()
