import React from 'react';
import * as api from './api.js';

class ValueViewer extends React.Component {
  constructor(props) {
    super(props);
    this.state = {sensor: {value: 0}};
    this.path = this.props.path
  }
  componentDidMount() {
    api.addOnUpdatedListener(this.path, this.onValueChanged.bind(this));
  }

  componentWillUnmount() {
    api.removeOnUpdatedListener(this.path, this.onValueChanged.bind(this));
  }

  onValueChanged(v) {
    this.setState({sensor: {value: v}});
  }

  render() {
    return (
      <span>
        {this.state.sensor.value.toFixed(this.props.fractionDigits)}
      </span>
    );
  }
}
export default ValueViewer;
