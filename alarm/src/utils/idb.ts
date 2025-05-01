const DB_NAME = 'alert-music-db'
const STORE_NAME = 'music'

export function saveFileToIDB(name: string, file: File): Promise<void> {
  return new Promise((resolve, reject) => {
    const open = indexedDB.open(DB_NAME, 1)
    open.onupgradeneeded = () => {
      open.result.createObjectStore(STORE_NAME)
    }
    open.onsuccess = () => {
      const db = open.result
      const tx = db.transaction(STORE_NAME, 'readwrite')
      tx.objectStore(STORE_NAME).put(file, name)
      tx.oncomplete = () => resolve()
      tx.onerror = () => reject(tx.error)
    }
    open.onerror = () => reject(open.error)
  })
}

export function getFileFromIDB(name: string): Promise<File | undefined> {
  return new Promise((resolve, reject) => {
    const open = indexedDB.open(DB_NAME, 1)
    open.onupgradeneeded = () => {
      open.result.createObjectStore(STORE_NAME)
    }
    open.onsuccess = () => {
      const db = open.result
      const tx = db.transaction(STORE_NAME, 'readonly')
      const req = tx.objectStore(STORE_NAME).get(name)
      req.onsuccess = () => resolve(req.result)
      req.onerror = () => reject(req.error)
    }
    open.onerror = () => reject(open.error)
  })
} 