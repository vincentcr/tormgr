/* global chrome */

/*
todo:
  - show feedback on add success/error
  - add regular webapp and firefox addon
*/

import * as Api from './api';
import {EventEmitter} from 'events';
import Session from './Session';

const TORRENT_MENU_RE = /^folders\.(.+)$/;
const LINK_CONTEXT = ['link'];

function onMenuClick(info, tab) {
  const {menuItemId,linkUrl} = info;
  const feedMatch = TORRENT_MENU_RE.exec(menuItemId);
  if (feedMatch) {
    const folderID = feedMatch[1];
    addFeedItem({folderID, link:linkUrl});
  } else {
    signin(tab);
  }
}

function addFeedItem({folderID, link}) {
  const item = {link, title:link};
  console.log('addFeedItem:', item);
  Api.Feeds.saveItem({folderID, item}).catch((err) => {
    console.log('error adding folder item:', err);
  });
}

function signin() {
  chrome.tabs.create({'url': chrome.extension.getURL('popup.html')});
}

function updateMenu() {
  console.log('updateMenu yo!!')
  chrome.contextMenus.removeAll();
  createMenu();
}

const apiProxyListener = new EventEmitter();
apiProxyListener.on('update', updateMenu);

chrome.runtime.onMessage.addListener((req, sender, callback) => {
  if (req.type === 'api') {
    receiveApiMessage(req.msg, callback);
    return true; //signal that we expect a callback
  }
});

function receiveApiMessage(msg, callback) {
  const {endpoint, method, params} = msg;
  const promise = Api[endpoint][method](...params);
  promise
    .then(res => {
      console.log(`api:${endpoint}.${method} successful`)
      apiProxyListener.emit('update');
      callback({res});
    })
    .catch(err => callback({err}))
  ;
}



function mkEndpointProxy(endpointName) {
  const proxy = {};
  const endpoint = Api[endpointName];
  Object.keys(endpoint).forEach((memberName) => {
    const member = endpoint[memberName];
    if (typeof member == 'function') {
      proxy[memberName] = (...params) => {
        return sendApiMessage(endpointName, memberName, params);
      };
    }
  });
  return proxy;
}



function sendApiMessage(endpoint, method, params) {
  const msg = {endpoint, method, params};
  return new Promise((success, fail) => {
    chrome.runtime.sendMessage({type:'api', msg}, ({err, res}) => {
      if (err) {
        fail(err);
      } else {
        success(res);
      }
    });

  });
}

export const ApiProxy = ['Users', 'Feeds'].reduce((proxy, name) => {
  proxy[name] = mkEndpointProxy(name);
  return proxy;
}, {});

function createMenu() {
  if (Session.isSignedIn()) {
    createFeedsMenu();
  } else {
    createSigninMenu();
  }
}

function createFeedsMenu() {
  Api.Feeds.getAll()
  .then((folders) => {
    folders.forEach((folder) => {
      const item = {title: 'add to ' + folder.name, contexts:LINK_CONTEXT, id: 'folders.' + folder.id};
      chrome.contextMenus.create(item);
    });
  }).catch((err) => {
    console.log('error creating folders menu:', err);
  });
}

function createSigninMenu() {
  chrome.contextMenus.create({title: 'Sign in to add folders', contexts:LINK_CONTEXT, id:'signin'});
}

chrome.contextMenus.onClicked.addListener(onMenuClick);
chrome.runtime.onInstalled.addListener(createMenu);
