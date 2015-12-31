// /* global require */
import 'babel/polyfill';
// import jquery from 'jquery';
// window.jQuery = jquery;
// require('bootstrap'); //if we use import the jQuery global will not be defined yet at import time
import Root from './components/root.jsx';
import * as Api from './api';

const defaultOptions = {
  api:Api,
  rootSelector:'#app-root',
};

export default function setupApp(options) {
  const {api, rootSelector} = Object.assign({}, defaultOptions, options);
  console.log('setupApp:', api, rootSelector);
  document.addEventListener('DOMContentLoaded', () =>  {
    Root.create({api, rootNode:document.querySelector(rootSelector)});
  });

  window.onerror = function(msg, url, line, col, err) {
    const timestamp = new Date().toISOString();
    console.log(`${timestamp} uncaught error: ${msg}   at:${line}:${col}\n`);
    if (err) {
      console.log(err);
    }
  };

}

window.setupApp = setupApp;
