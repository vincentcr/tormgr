
import React, {PropTypes} from 'react';
import { Link } from 'react-router';
import { saveFeedItem } from '../actions';

export const APP_TITLE = 'TORMGR';

export default class Home extends React.Component {

  constructor(...args) {
    super(...args);
    this.state = {};
  }

  static propTypes = {
    dispatch: PropTypes.func.isRequired,
    folders: PropTypes.object.isRequired,
  }

  render() {
    const {folders} = this.props;

    const quickAdd = this.renderQuickAddTorrent(folders);
    const summary = this.renderSummary(folders);

    return (
      <div>
        <div className='row titlebar'>
          <div className='col-md12'>
            <h2 className='title'>{APP_TITLE}</h2>
          </div>
        </div>
        <div className='row'>
          <div className='col-md-12'>
            {summary}
          </div>
        </div>
        <div className='row'>
          <div className='col-md-12'>
            {quickAdd}
          </div>
        </div>
      </div>
    );

  }

  renderQuickAddTorrent(folders) {
    const folderSelect = this.renderQuickAddFoldersSelection(folders);
    const valid = this.state.folderID != null && this.state.link != null;

    return (
      <form className='form-inline' onSubmit={(e) => this.handleAddItem(e)}>
        <div className='form-group'>
          <label htmlFor='quick-add-url' className='sr-only'>Email</label> {' '}
          <input type='text' required={true} onChange={(e) => this.handleOnChange(e)} value={this.state.link} className='form-control' id='quick-add-url' placeholder='url or info hash' />
          <input type='hidden' required={true} value={this.state.folderID} />
        </div>
        {' '}
        <div className='form-group'>
          {folderSelect}
        </div>

        {' '}
        <button type='submit' className='btn btn-default' disabled={!valid}>
          <span className='glyphicon glyphicon-plus' aria-hidden='true'></span>
          {' '}
          add torrent
        </button>
      </form>

    );
  }

  handleOnChange(e) {
    this.setState({ link : e.target.value });
  }

  handleAddItem(e) {
    e.preventDefault();
    const {link,folderID} = this.state;
    const {folders, dispatch} = this.props;
    const item = {link, title:link};
    dispatch(saveFeedItem(folders[folderID], item));
  }

  renderQuickAddFoldersSelection(folders) {

    const {folderID} = this.state;

    const selectFolder = (e, folder) => {
      e.preventDefault();
      this.setState({folderID:folder.id});
    };

    const items = Object.values(folders).map(folder =>
      <li key={folder.id}>
        <a href='#' onClick={(e) => selectFolder(e, folder)}>
          {folder.name}
        </a>
      </li>
    );

    const name = folderID != null ? folders[folderID].name : '[Select folder]';

    return (
      <div className='folder-select dropdown'>
        <button className='btn btn-default dropdown-toggle' type='button' id='quick-add-select-folder'
          data-toggle='dropdown' aria-haspopup='true' aria-expanded='true'>
          {name + ' '}
          <span className='caret'></span>
        </button>
        <ul className='dropdown-menu' aria-labelledby='quick-add-select-folder'>
          {items}
        </ul>
      </div>
    );
  }


  renderSummary(folders) {
    const items = Object.values(folders).map((folder,idx) => this.renderSummaryFeed({folder,idx}));
    return (
      <table className='folders-summary table table-condensed'>
        {items}
      </table>
    );
  }

  renderSummaryFeed({idx, folder}) {
    function onCopy(e) {
      e.preventDefault();
      copyTextToClipboard(folder.link);
    }
    return (
      <tr key={folder.id}>
        <td>
          {idx + 1}.
        </td>
        <td>
          {folder.name}
        </td>
        <td>
          <Link className='btn btn-default btn-sm folder-link' to={`/folders/${folder.id}`} title='edit'>
            <span className='glyphicon glyphicon-edit' aria-hidden="true"></span>
          </Link>
        </td>
        <td>
          <a className='btn btn-default btn-sm clipboard' onClick={onCopy} href='#' aria-label='Copy To clipboard' title='Copy rss link to clipboard'>
            <span className='glyphicon glyphicon-copy' aria-hidden="true"></span>
          </a>
        </td>
      </tr>
    );
  }

}

function copyTextToClipboard(text) {
  var textArea = document.createElement('textarea');

  // Place in top-left corner of screen regardless of scroll position.
  textArea.style.position = 'fixed';
  textArea.style.top = 0;
  textArea.style.left = 0;

  // Ensure it has a small width and height. Setting to 1px / 1em
  // doesn't work as this gives a negative w/h on some browsers.
  textArea.style.width = '2em';
  textArea.style.height = '2em';

  // We don't need padding, reducing the size if it does flash render.
  textArea.style.padding = 0;

  // Clean up any borders.
  textArea.style.border = 'none';
  textArea.style.outline = 'none';
  textArea.style.boxShadow = 'none';

  // Avoid flash of white box if rendered for any reason.
  textArea.style.background = 'transparent';


  textArea.value = text;

  document.body.appendChild(textArea);

  textArea.select();

  try {
    var successful = document.execCommand('copy');
    var msg = successful ? 'successful' : 'unsuccessful';
    console.log('Copying text command was ' + msg);
  } catch (err) {
    console.log('Oops, unable to copy');
  }

  document.body.removeChild(textArea);
}
