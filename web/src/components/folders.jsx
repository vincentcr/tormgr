import React, {PropTypes} from 'react';
import { Link } from 'react-router';
import { connect } from 'react-redux';
import { fetchFoldersIfNeeded, createFolder } from '../actions';
import Session from '../session';
import Home from './home.jsx';

const NEW_FOLDER_PFX = 'New Folder ';

@connect(state => state.folder)
export default class Folders extends React.Component {

  static propTypes = {
    children: PropTypes.any,
    dispatch: PropTypes.func.isRequired,
  }

  static contextTypes = {
    history: PropTypes.object.isRequired,
  }

  componentDidMount() {
    const { dispatch } = this.props;
    dispatch(fetchFoldersIfNeeded());
  }

  render() {
    const { folders, isFetching, children, dispatch, err } = this.props;
    const folderNodes = Object.values(folders).map(folder => this.renderFolder(folder));
    const userMenu = this.renderUserMenu();

    const content = children != null ? children : <Home folders={folders} dispatch={dispatch} />;

    return (
      <div>
        <div className='container-fluid'>
          <div className='row'>
            <div className='col-sm-3 col-md-2 sidebar'>
              {userMenu}

              {isFetching &&
                <div className='loading'>Loading...</div>
              }

              {!isFetching &&
                <ul className='nav nav-sidebar navbar-nav folders-sidebar'>

                  <li className={this.props.params.folderID == null ? 'active' : ''}>
                    <Link className='folder-link' to='/folders'>
                      <span className='glyphicon glyphicon-home' aria-hidden='true'></span>
                      {' '}
                      home
                    </Link>
                  </li>

                  {folderNodes}
                  <li className='folder-add-new'>
                    <hr />
                    <a href='#' onClick={e => this.handleCreateFolder(e)}>
                      <span className='glyphicon glyphicon-plus' aria-hidden='true'></span>
                      {' '}
                      add new folder
                    </a>
                  </li>
                </ul>
              }
              </div>
          </div>
        </div>

        <div className="col-sm-9 col-sm-offset-3 col-md-10 col-md-offset-2 main">
          {content}

          {err &&
            <span className='error'>
              {err.message || err.toString()}
            </span>
          }

        </div>
      </div>
    );
  }

  renderFolder(folder) {

    const selected = this.props.params.folderID === folder.id;
    const className = (selected) ? 'active' : undefined;

    return (
      <li key={folder.id} className={className}>
        <Link className='folder-link' to={`/folders/${folder.id}`}>
          <span className='glyphicon glyphicon-th-list' aria-hidden='true'></span>
          {' '}
          {folder.name}
        </Link>
      </li>
    );
  }

  renderUserMenu() {
    const user = Session.get('user');

    return (
      <ul className='nav nav-sidebar'>
        <li className="dropdown">
          <a href="#" className="dropdown-toggle" data-toggle="dropdown" role="button" aria-haspopup="true" aria-expanded="false">
            {user.email}
            <span className="caret"></span>
          </a>
          <ul className="dropdown-menu">
            <li>
              <a className='signout' onClick={e => this.handleSignout(e)}>
                <span className='glyphicon glyphicon-log-out' aria-hidden="true"></span>
                {' '}
                signout
              </a>
            </li>
          </ul>
        </li>
        <li role='separator' className='divider'><hr /></li>
      </ul>

    );
  }

  handleSignout() {
    const {api} = this.props;
    api.Users.signout();
  }

  handleCreateFolder(e) {
    e.preventDefault();
    const {dispatch} = this.props;
    const name = this.defaultNewName();
    dispatch(createFolder(name));
  }

  defaultNewName() {
    const {folders} = this.props;
    const idxExtractor = new RegExp(`^${NEW_FOLDER_PFX}(\\d+)$`);
    const lastIdx = Math.max(0, ...Object.values(folders).map(({name}) => {
      const match = idxExtractor.exec(name);
      if (match) {
        return parseInt(match[1]);
      } else {
        return 0;
      }
    }));
    const name = NEW_FOLDER_PFX + (lastIdx + 1);
    return name;
  }
}
