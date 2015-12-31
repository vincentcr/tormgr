import Session from './session';
import history from './history';
import 'whatwg-fetch';
import config from './config';

const MIME_JSON = 'application/json';
const DEFAULT_HEADERS = Object.freeze({
  'accept': MIME_JSON,
  'content-type': MIME_JSON,
});

const baseUrl = config.api.baseUrl;
console.log('api:baseUrl:', baseUrl);

export class Api {

  get(url, opts) {
    return this.send(url, {...opts, method: 'GET'});
  }

  post(url, opts) {
    return this.send(url, {...opts, method: 'POST'});
  }

  put(url, opts) {
    return this.send(url, {...opts, method: 'PUT'});
  }

  delete(url, opts) {
    return this.send(url, {...opts, method: 'DELETE'});
  }

  send(url, opts) {
    url = this._normalizeUrl(url);
    this._completeHeaders(opts);
    this._encodeBody(opts);
    const res = this._exec(url, opts);
    return res;
  }

  _normalizeUrl(url) {
    if(!/^https?:/.test(url)) {
      return baseUrl + url;
    } else {
      return url;
    }
  }

  _completeHeaders(opts) {
    if (opts.headers == null) {
      opts.headers = {};
    }
    opts.headers = Object.assign({}, DEFAULT_HEADERS, opts.headers);
    this._authorize(opts);
  }

  _authorize(opts) {
    if (opts.auth != null) {
      this._authorizeWith(opts, opts.auth);
    } else {
      const token = Session.get('token');
      if (token != null) {
        this._authorizeWith(opts, {scheme: 'token', creds: token});
      }
    }

    delete opts.auth;
  }

  _authorizeWith(opts, {scheme, creds}) {
    if (opts.headers.authorization == null) {
      if (scheme === 'token') {
        opts.headers.authorization = `Token token="${creds}"`;
      } else if (scheme === 'basic') {
        const encoded = btoa(creds.email + ':' + creds.password);
        opts.headers.authorization = `Basic ${encoded}`;
      }
    }
  }

  _encodeBody(opts) {
    if (isJSON(opts.headers['content-type']) && typeof opts.data === 'object' && opts.body == null) {
      opts.body = JSON.stringify(opts.data);
      delete opts.data;
    }
  }

  _exec(url, opts) {
    return fetch(url, opts).then((res) => {
        const status = res.status;
        if (status >= 400) {
          const err = Object.assign(new Error(`Unexpected status ${res.status}`), {res, status});
          throw err;
        } else {
          return res;
        }
      })
      .catch((err) => {
        if (err.status === 401) {
          if (Session.isSignedIn()) {
            console.log('invalid token, signout');
            Users.signout();
          }
        }
        console.log('request failed', err, {url, opts});
        throw err;
      })
      ;
  }

}

Api.create = function(routes) {
  const api = new Api();
  routes._api = api;
  return routes;
};



function isJSON(contentType) {
  if (contentType) {
    const mime = /^(.+?)(\;.*)?$/.exec(contentType)[1];
    return MIME_JSON === mime;
  }
}


export const Users = Api.create({
  signup(creds) {
    return this._api.post('/users', { data: creds  })
    .then(res => res.json())
    .then(userData => Session.set(userData))
    ;
  },

  signin(creds) {
    return this._api.post('/users/tokens', {
      auth: {
        scheme: 'basic',
        creds: creds,
      },
    })
    .then(res => res.json())
    .then(userData => Session.set(userData))
    ;
  },

  signout() {
    const token = Session.get('token');
    function finalize() {
      Session.clear();
      setTimeout(() => history.replaceState(null, '/signin'));
    }

    if (token != null) {
      return this._api.delete(`/users/tokens/${token}`).then(finalize).catch(finalize);
    } else {
      return Promise.reject(new Error('not signed in'));
    }
  },

});

export const Folders = Api.create({

  getAll() {
    return this._api.get('/folders').then(res => res.json());
  },

  get(folderID) {
    return this._api.get(`/folders/${folderID}`).then(res => res.json());
  },

  create(name) {
    const folder = {name};
    return this._api.post('/folders', {data:folder}).then(res => res.json());
  },

  update(folder) {
    const data = {name: folder.name};
    return this._api.put(`/folders/${folder.id}`, {data});
  },

  delete(folder) {
    return this._api.delete(`/folders/${folder.id}`);
  },

});

export const Torrents = Api.create({

  getByFolder(folder) {
    return this._api.get(`/folders/${folder}/torrents`).then(res => res.json());
  },

  get(torrentID) {
    return this._api.get(`/torrents/${torrentID}`).then(res => res.json());
  },

  create({folder, urlOrInfoHash}) {
    const data = {folder, urlOrInfoHash};
    return this._api.post('/torrents', {data}).then(res => res.json());
  },

  update(torrent) {
    const data = {folder:torrent.folder};
    return this._api.put(`/torrents/${torrent.id}`, {data});
  },

  deleteItem(torrent) {
    return this._api.delete(`/torrents/${torrent.id}`);
  },

});
