import React from 'react';
import Grid from '@material-ui/core/Grid';
import Value from "./Value.js";
import Slider from "./Slider.js";
import Switch from "./Switch.js";
import Light from "./Light.js";
import './App.css';

function App() {
    return (
	<div>
	    <h1>
	      Devices
	    </h1>
	    <Grid container alignItems="center" spacing={8}>
	    {["device1", "device2"].map(deviceID => (
		<Grid item xs key={deviceID}>
		  <h2>
  		    {deviceID}
   		  </h2>
		  <p>Light</p>
		  <Light deviceID={deviceID} ID="light" />
		  <p>Value</p>
		  <Value deviceID={deviceID} ID="value" />
		  <p>Switch</p>
		  <Switch deviceID={deviceID} ID="switch" />
		  <p>Slider</p>
		  <Slider deviceID={deviceID} ID="slider" />
	      </Grid>
	    ))}
	    </Grid>
	</div>
  );
}

export default App;
