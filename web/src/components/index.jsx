import React, {PropTypes} from 'react';
import { connect } from 'react-redux';

@connect(state => state.asyncState)
export default class Index extends React.Component {

  static propTypes = {
    children: PropTypes.any,
  }

  render() {
    const {children, inProgress} = this.props;
    const inProgressClass = inProgress ? 'active' : '';
    return (
      <div className='app'>
        <div id='in-progress' className={inProgressClass}>
          <div className="throbber-loader">
            Loading
          </div>
        </div>
        {children}
      </div>
    );
  }
}
