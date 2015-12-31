import history from './history';
import {
  ASYNC_BEGIN,
  ASYNC_COMPLETE,
  FOLDERS_FETCH_INVALIDATE,
  FOLDERS_FETCH_COMPLETE,
  FOLDERS_UPDATE,
  FOLDERS_DELETE,
  FOLDER_FETCH_INVALIDATE,
  FOLDER_FETCH_COMPLETE,
  FOLDER_CREATE,
  FOLDER_SELECT,
  TORRENT_DESELECT,
  TORRENT_UPDATE,
  TORRENT_DELETE,
} from './stores';

let api = null;
export function setApi(newApi) {
  api = newApi;
}

function asyncBegin() {
  return { type:ASYNC_BEGIN };
}

function asyncComplete(res) {
  return { type:ASYNC_COMPLETE, res};
}

export function invalidateFolders() {
  return { type:FOLDERS_FETCH_INVALIDATE };
}

function fetchIfNeeded(stateKey, fetch) {
  return (dispatch, getState) => {
    const state = getState();
    const subState = state[stateKey];
    if (shouldFetch(subState)) {
      dispatch(asyncBegin());
      fetch(dispatch)
        .then(res => dispatch(asyncComplete(res)) )
        .catch(err => dispatch(asyncComplete({err})))
      ;
    }
  };
}

function shouldFetch(state) {
  if (state.isFetching) {
    return false;
  } else {
    return state.didInvalidate;
  }
}

export function fetchFoldersIfNeeded() {
  return fetchIfNeeded('folders', (dispatch) => {
    return api.Folders.getAll()
      .then(folders => { return {folders}; })
      .catch(err => { return {err, folders:[]}; })
      .then(res => dispatch(completeFetchFolders(res)))
      ;
  });
}

function completeFetchFolders({folders,err}) {
  return { type:FOLDERS_FETCH_COMPLETE, folders, err};
}

export function fetchFolderIfNeeded(folder) {
  const getState = (fullState) => fullState.folder
  return fetchIfNeeded('folder', (dispatch) => {
    return api.Torrents.getByFolder(folder)
      .then(torrents => { return {torrents}; })
      .catch(err => { return {err, torrents:[]}; })
      .then(res => dispatch(completeFetchFolder(res)))
      ;
  });
}

function completeFetchFolder({folders,err}) {
  return { type:FOLDER_FETCH_COMPLETE, folders, err};
}


export function invalidateFolder(folder) {
  return { type:FOLDERS_FETCH_INVALIDATE, folder };
}

export function fetchCurrentFeedIfNeeded(folderID) {
  return (dispatch, getState) => {
    dispatch(fetchFeedsIfNeeded(() => {
      const folder = getState().folder.folders[folderID];
      if (folder == null) {
        history.replaceState(null, '/folders');
      } else {
        dispatch(selectFolder(folder));
      }
    }));
  };
}

export function createFeed(title) {
  return (dispatch) => {
    return api.Feeds.create(title).then((folder) => {
      console.log('folder created:', folder);
      dispatch({type:FOLDER_CREATE, folder});
      dispatch({type:FOLDERS_UPDATE, folder});
      history.replaceState(null, `/folders/${folder.id}`);
    });
  };
}

export function deleteFeed(folder) {
  return (dispatch, getState) => {
    const {folders} = getState().folder;
    const adjFeed = findAdjacentFeed({folder, folders});

    return api.Feeds.delete(folder).then(() => {
      dispatch({type:FOLDERS_DELETE, folder});
      if (adjFeed != null) {
        dispatch(selectFolder(adjFeed));
        history.replaceState(null, `/folders/${adjFeed.id}`);
      } else {
        dispatch({type:TORRENT_DESELECT});
        history.replaceState(null, '/folders');
      }
    });
  };
}

function findAdjacentFeed({folder, folders}) {
  const feedIDs = Object.keys(folders);
  const curIdx = feedIDs.findIndex(id => id === folder.id);
  const adjIdx = (curIdx === 0) ? 1 : curIdx - 1;
  return folders[feedIDs[adjIdx]];
}

export function selectFolder(folder) {
  return {type:FOLDER_SELECT, folder};
}

export function saveFeed(folder) {
  return (dispatch, getState) => {
    const origFeed = getState().folder.folder;
    updateFeed({dispatch, folder});
    dispatch(asyncBegin());
    return api.Feeds.save(folder)
      .catch(err => {
        console.log('error', err);
        updateFeed({dispatch, folder:origFeed});
        return {err};
      })
      .then(res => dispatch(asyncComplete(res)));
  };
}

function updateFeed({dispatch, folder}) {
  dispatch({type:TORRENT_UPDATE, folder});
  dispatch({type:FOLDERS_UPDATE, folder});
}

export function saveFeedItem(folder, item, done = () => null) {
  return (dispatch) => {
    const origItem = folder.items.find(i => i.id === item.id);
    dispatch(asyncBegin());
    return api.Feeds.saveItem({folderID:folder.id, item})
      .then(item => {
        updateFeedItem({dispatch, folder, item});
        return item;
      })
      .then(done)
      .catch(err => {
        console.log('error', err);
        if (origItem != null) {
          updateFeedItem({dispatch, folder, item:origItem});
        }
        return {err};
      })
      .then(res => dispatch(asyncComplete(res)));
  };
}

function updateFeedItem({dispatch, folder, item}) {
      const items = folder.items
        .filter(i => i.id !== item.id)
        .concat(item);
      const updatedFeed = {...folder, items};
      updateFeed({dispatch, folder:updatedFeed});
}

export function deleteFeedItem(folder, item) {
  return (dispatch) => {
    return api.Feeds.deleteItem({folder, item}).then(() => {
      dispatch({type:TORRENT_DELETE, folder, item});
      history.replaceState(null, `/folders/${folder.id}`);
    });
  };
}
