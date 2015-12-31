const STORAGE_ROOT = 'myfolders.session.';

export class Session {

  constructor() {
    this._data = {};
    for (let fqKey in localStorage) {
      if (fqKey.startsWith(STORAGE_ROOT)) {
        const key = fqKey.substr(STORAGE_ROOT.length);
        const val = JSON.parse(localStorage[fqKey]);
        this._data[key] = val;
      }
    }
  }

  isSignedIn() {
    return this.get('token') != null;    
  }

  get(key) {
    return this._data[key];
  }

  set(data) {
    for (let key in data) {
      const val = data[key];
      this._data[key] = val;
      localStorage[this._fqKey(key)] = JSON.stringify(val);
    }
  }

  remove(key) {
    console.log(`session:delete:${key}`);
    delete this._data[key];
    delete localStorage[this._fqKey(key)];
  }

  clear() {
    console.log('session:clear');
    for (let [key] of this) {
      this.remove(key);
    }
  }

  [Symbol.iterator]() {
    return Object.entries(this._data)[Symbol.iterator]();
  }

  _fqKey(key) {
    return STORAGE_ROOT + key;
  }
}

export default new Session();
