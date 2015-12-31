import React, {PropTypes} from 'react';
import { connect } from 'react-redux';
import classNames from 'classnames';
import history from '../history';
import {
  fetchTorrentsIfNeeded,
  saveFeed,
  deleteFeed,
} from '../actions';

@connect(state => state.folder)
export default class Feed extends React.Component {

  constructor() {
   super();
   this.state = { };
  }

  static propTypes = {
    children: PropTypes.any,
    dispatch: PropTypes.func.isRequired,
  }

  componentDidMount() {
    this._loadFolder(this.props);
  }

  componentWillReceiveProps(nextProps) {
    this._loadFolder(nextProps);
  }

  _loadFolder(props) {
    const newName = props.params.name;
    const {name} = this.state;
    if (newName !== name) {
      const {dispatch} = this.props;
      this.setState({name:newName});
      dispatch(fetchTorrentsIfNeeded(newName));
    }
    const folder = props.folder ? {...props.folder} : null;
    this.setState({folder, isModified:false});
  }

  render() {
    const cssClasses = classNames({
      'edit-mode' : this.props.isEditing,
    });
    const {folder} = this.state;
    const {params, children} = this.props;

    return (
      <div className={'folder ' + cssClasses}>
        {!folder &&
          <div className='loading'>Loading...</div>
        }
        {folder &&
          <div>
            <div className='row titlebar'>
              <div className='col-md12'>
                {this.renderTitleBar(folder)}
              </div>
            </div>
            <div className='row'>
              <div className='col-md-6'>
                <TorrentsTable ref='itemList' params={params} folder={folder} />
                <div className='folder-item-add-new'>
                  <button onClick={(e) => this.handleAddItem(e)} className='btn btn-default'>Add Feed Item</button>
                </div>
              </div>
              <div className='col-md-6'>
                {children}
              </div>
            </div>
          </div>
        }
      </div>
    );
  }

  renderTitleBar(folder) {
    return (
      <div>
        {this.renderTitle(folder)}
        {' '}
        <div className='dropdown folder-menu'>
          <button className="btn btn-default dropdown-toggle" type="button" id="dropdownMenu1" data-toggle="dropdown" aria-haspopup="true" aria-expanded="true">
            <span className="glyphicon glyphicon-cog" aria-hidden="true"></span>
          </button>
          <ul className="dropdown-menu" aria-labelledby="dropdownMenu1">
            <li><a href={folder.link} target={'rss' + folder.id}>
                <span className='glyphicon glyphicon-th-list' aria-hidden="true"></span>
                {' '}
                view rss
            </a></li>
            <li><a href="#" onClick={e => this.handleDelete(e)}>
                <span className='glyphicon glyphicon-trash' aria-hidden="true"></span>
                {' '}
                delete folder
            </a></li>
          </ul>
        </div>
      </div>
    );
  }

  renderTitle(folder) {
    const handleOnChange = this.handleOnChange.bind(this, 'title');

    function handleOnBlur(e) {
      //using this hidden button trick, we can trigger browser form validation
      const button = e.target.form.querySelector('button');
      setTimeout(() => button.click(), 0);
    }

    return (
      <h2 className='title'>
        <form onSubmit={e => this.handleSave(e)}>
          <input type='text' className="form-control" placeholder='title' required='required'
            value={folder.name} size={folder.name.length}
            onChange={handleOnChange} onBlur={handleOnBlur}
          />
        <button style={{display:'none'}} type='submit' />
      </form>
      </h2>
    );
  }

  handleSave(e) {
    e.preventDefault();
    const {folder, isModified} = this.state;
    if (isModified) {
      this.props.dispatch(saveFeed(folder));
    }
  }

  handleOnChange(field, e) {
    const folder = {...this.state.folder, [field] : e.target.value };
    this.setState({ isModified: true, folder:folder });
  }

  handleDelete(e) {
    const {folder} = this.state;
    const {dispatch} = this.props;
    e.preventDefault();
    if (window.confirm(`Delete folder ${folder.name}? This action cannot be undone.`)) {
      dispatch(deleteFeed(folder));
    }
  }

  handleAddItem(e) {
    e.preventDefault();
    const {folder} = this.props;
    showItemDetails({folderID:folder.id, itemID:'new'});
  }
}

class TorrentsTable extends React.Component {

  getItems() {
    const items = Object.keys(this.refs)
      .filter(key => key.startsWith('item_'))
      .map(key => this.refs[key]);
    return items;
  }

  render() {
    const {folder, params} = this.props;
    const nodes = folder.items.map((torrent, idx) => React.createElement(TorrentRow, {
        torrent,
        params,
        idx: idx,
        key: torrent.id,
    }));
    const showAsEmpty = nodes.length === 0;
    return (
      <table className='folder-item-table table table-condensed' style={{width: 'auto'}}>
        <thead>
          <tr>
            <th></th>
            <th> title </th>
            <th> url </th>
          </tr>
        </thead>
        <tbody>
          {nodes}
          {showAsEmpty &&
            <tr className='empty-folder'>
              <td colSpan='4'>(empty)</td>
            </tr>
          }
        </tbody>
      </table>
      );
  }
}

class TorrentRow extends React.Component {

  constructor(props) {
    super(props);
    const {item} = this.props;
    this.state = {...item};
  }

  render() {
    const {idx,torrent,params} = this.props;
    const {folderID, torrentID} = params;
    const cssClasses = classNames({
      'selected' : torrentID === torrent.id,
    });
    const onClick = () => showItemDetails({folderID, torrentID:torrent.id});
    return (
      <tr className={cssClasses} onClick={onClick}>
        <td>{idx + 1}.</td>
        <td>{torrent.title}</td>
        <td>
          {torrent.url}
          {' '}
          <a href={torrent.link} target={torrent.id} className='torrent-link'>
              <span className="glyphicon glyphicon-new-window" aria-hidden="true"></span>
          </a>
        </td>
        <td>{torrent.description}</td>
      </tr>
    );
  }
}

function showItemDetails({folderID, torrentID}) {
  const url = `/folders/${folderID}/${torrentID}`;
  history.replaceState(null, url);
}
