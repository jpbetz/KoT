import React from 'react';
import Slider from '@material-ui/core/Slider';
import * as api from './api.js';

class SensorSlider extends React.Component {
    constructor(props) {
	super(props);
	this.state = {sensor: {value: 0}};
	this.path = this.props.deviceID + "." + this.props.ID
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
	this.setState({sensor: {value: value}}) // pre-emptively update UI
	api.setOutput(this.props.deviceID, this.props.ID, value)
    }

    render() {
	return (
		<Slider onChange={(event, value) => this.onSliderChanged(event, value)} value={this.state.sensor.value} />
	);
    }

}
export default SensorSlider;
