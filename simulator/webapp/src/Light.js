import React from 'react';
import WbIncandescentIcon from '@material-ui/icons/WbIncandescent';
import * as api from './api.js';

class Light extends React.Component {
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
      let color = v > 0 ? 'error' : 'disabled';
      return (
        <WbIncandescentIcon color={color} fontSize={'large'} />
      );
    }

}
export default Light;
