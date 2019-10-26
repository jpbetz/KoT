import React from 'react';
import Switch from '@material-ui/core/Switch';
import * as api from './api.js';

class SensorSwitch extends React.Component {
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

    onSwitchChanged(event) {
	var value = event.target.checked ? 1 : 0;
	this.setState({sensor: {value: value}}) // pre-emptively update UI
	api.setOutput(this.props.deviceID, this.props.ID, value)
    }

    render() {
	return (
		<Switch onChange={(event, value) => this.onSwitchChanged(event, value)} checked={this.state.sensor.value ? true : false} />
	);
    }

}
export default SensorSwitch;
