import React from 'react';
import Slider from '@material-ui/core/Slider';
import * as api from './api.js';

class SensorSlider extends React.Component {
  constructor(props) {
    super(props);
    this.state = {sensor: {value: 0}};
    this.path = this.props.path;
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

  onSliderChanged(event, value) {
    let num = value.toString();
    this.setState({sensor: {value: value}}); // pre-emptively update UI
    if(this.props.input) {
      api.setInput(this.path, num)
    } else {
      api.setOutput(this.path, num)
    }
  }

  render() {
    let {deviceID, ID, input, ...other} = this.props;
    let v = Number.parseFloat(this.state.sensor.value);
    return (
      <Slider {...other} onChange={(event, value) => this.onSliderChanged(event, value)} value={v} />
    );
  }
}


export default SensorSlider;
