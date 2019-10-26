import React from 'react';
import WbIncandescentIcon from '@material-ui/icons/WbIncandescent';
import * as api from './api.js';

class Light extends React.Component {
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

    render() {
	var color = this.state.sensor.value > 0 ? 'primary' : 'disabled';
	return (
		<WbIncandescentIcon color={color} fontSize={'large'} />
	);
    }

}
export default Light;
