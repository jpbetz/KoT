import React from 'react';
import * as api from './api.js';

class ValueViewer extends React.Component {
  constructor(props) {
    super(props);
    this.state = {sensor: {value: 0}};
    this.path = this.props.path;
    this.onChangedHandler = this.onValueChanged.bind(this);
  }
  componentDidMount() {
    api.addOnUpdatedListener(this.path, this.onChangedHandler);
  }

  componentWillUnmount() {
    api.removeOnUpdatedListener(this.path, this.onChangedHandler);
  }

  onValueChanged(v) {
    this.setState({sensor: {value: v}});
  }

  render() {
    let v = Number.parseFloat(this.state.sensor.value);
    return (
      <span>
        {v.toFixed(this.props.fractionDigits)}
      </span>
    );
  }
}

export default ValueViewer;
