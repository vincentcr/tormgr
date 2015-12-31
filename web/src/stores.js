import { createStore, applyMiddleware, combineReducers } from 'redux';
import thunkMiddleware from 'redux-thunk';
import createLoggerMiddleware from 'redux-logger';

export const ASYNC_BEGIN = 'ASYNC_BEGIN';
export const ASYNC_COMPLETE = 'ASYNC_COMPLETE';

export const FOLDERS_FETCH_INVALIDATE = 'FOLDERS_FETCH_INVALIDATE';
export const FOLDERS_FETCH_COMPLETE = 'FOLDERS_FETCH_COMPLETE';
export const FOLDERS_UPDATE = 'FOLDERS_UPDATE';
export const FOLDERS_DELETE = 'FOLDERS_DELETE';

export const FOLDER_FETCH_INVALIDATE = 'FOLDER_FETCH_INVALIDATE';
export const FOLDER_FETCH_COMPLETE = 'FOLDER_FETCH_COMPLETE';
export const FOLDER_CREATE = 'FOLDER_CREATE';
export const FOLDER_SELECT = 'FOLDER_SELECT';
export const FOLDER_UPDATE = 'FOLDER_UPDATE';

export const TORRENT_DELETE = 'TORRENT_DELETE';

const INITIAL_STATE = Object.freeze({
  asyncState: {
    inProgress:false,
  },
  folders: {
    didInvalidate: true,
    folders: {},
  },
  torrents: {
    torrents : {},
  },
  folder: {
    current: null,
  },
});

function asyncState(state = INITIAL_STATE.asyncState, action) {
  switch (action.type) {
    case ASYNC_BEGIN:
      return {...state, inProgress:true};
    case ASYNC_COMPLETE:
      return {...state, inProgress:false};
    default:
      return state;
  }
}

function folders(state = INITIAL_STATE.folders, action) {
  console.log(`store:folders:${action.type}`, {state, action});
  switch(action.type) {
    case FOLDERS_FETCH_INVALIDATE:
      return {...state, didInvalidate:true};
    case FOLDERS_FETCH_COMPLETE:
      return {...state, err:action.err, folders:toFolderMap(action.folders), didInvalidate:false};
    case FOLDERS_UPDATE: {
      const {folder} = action;
      const folders = {...state.folders, [folder.name]:folder};
      return {...state, folders:folders};
    }
    case FOLDERS_DELETE: {
      const {folder} = action;
      const folders = {...state.folders};
      delete folders[folder.name];
      return {...state, folders:folders};
    }
    default:
      return state;
  }
}

function toFolderMap(folder) {
  const folders = folder.reduce((folders, folder) => {
    folders[folder.name] = folder;
    return folders;
  }, {});
  return folders;
}

function folder(state = INITIAL_STATE.folder, action) {
  console.log(`store:folder:${action.type}`, {state, action});
  switch(action.type) {
    case FOLDER_FETCH_INVALIDATE:
      return {...state, didInvalidate:true};
    case FOLDER_FETCH_COMPLETE:
      return {...state, err:action.err, torrents:action.torrents, didInvalidate:false};
    case FOLDER_CREATE:
      return {...state, folder:action.folder};
    case FOLDER_SELECT:
      return {...state, folder:action.folder};
    case FOLDER_UPDATE:
      return {...state, folder:action.folder};
    case TORRENT_DELETE: {
      const {item, folder} = action;
      const items = folder.items.filter(i => i != item);
      const updatedFeed = {...folder, items};
      return {...state, folder:updatedFeed};
    }
    default:
      return state;
  }
}

function torrents(state = INITIAL_STATE.torrents, action) {
  console.log(`store:torrents:${action.type}`, {state, action});
}

function torrentsByFolder(state = INITIAL_STATE.torrents, action) {
  console.log(`store:torrents:${action.type}`, {state, action});
  const torrentsState = torrents(state[action.folder], action);
  return {...state, [action.folder]: torrentsState  };
}


const rootReducer = combineReducers({folders, folder, torrentsByFolder, asyncState});

const createStoreWithMiddleware = applyMiddleware(
  thunkMiddleware,
  createLoggerMiddleware()
)(createStore);

export default function configureStore(initialState) {
  return createStoreWithMiddleware(rootReducer, initialState);
}
