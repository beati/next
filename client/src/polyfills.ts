import 'core-js/es6';
import 'reflect-metadata';
import 'webrtc-adapter/out/adapter';
require('zone.js/dist/zone');

if (process.env.ENV === 'production') {
	// Production
} else {
	// Development
	Error['stackTraceLimit'] = Infinity;
	require('zone.js/dist/long-stack-trace-zone');
}
