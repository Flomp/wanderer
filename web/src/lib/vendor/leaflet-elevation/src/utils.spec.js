/**
 * src/utils.js
 */

import { suite } from 'uvu';
import * as assert from 'uvu/assert';
import '../test/setup/jsdom.js'
import { iAvg, iMin, iMax, iSum } from "../src/utils.js";

const toFixed = (n) => +n.toFixed(2);

const test = suite('src/utils.js');

test('iAvg()', () => {
    let avg;
    avg = iAvg(100, undefined, 1); assert.is(toFixed(avg), 100);    // average for [100] is 100
    avg = iAvg(100, avg, 2);       assert.is(toFixed(avg), 100);    // average for [100, 100] is 100
    avg = iAvg(200, avg, 3);       assert.is(toFixed(avg), 133.33); // average for [100, 100, 200] is 133.33
    avg = iAvg(200, avg, 4);       assert.is(toFixed(avg), 150);    // average for [100, 100, 200, 200] is 150
    avg = iAvg(NaN, avg, 5);       assert.ok(isNaN(avg));           // average for [100, 100, 200, 200, NaN] is NaN
});

test('iMin()', () => {
    let min;
    min = iMin(100, undefined); assert.is(toFixed(min), 100);  // min for [100] is 100
    min = iMin(NaN, min);       assert.is(toFixed(min), 100);  // min for [100, NaN] is 100
    min = iMin(0, min);         assert.is(toFixed(min), 0);    // min for [100, NaN, 0] is 100
    min = iMin(-200, min);      assert.is(toFixed(min), -200); // min for [100, NaN, 0, -200] is -200
    min = iMin(200, min);       assert.is(toFixed(min), -200); // min for [100, NaN, -100, -200, 200] is -200
});

test('iMax()', () => {
    let max;
    max = iMax(100, undefined); assert.is(toFixed(max), 100); // max for [100] is 100
    max = iMax(NaN, max);       assert.is(toFixed(max), 100); // max for [100, NaN] is 100
    max = iMax(0, max);         assert.is(toFixed(max), 100); // max for [100, NaN, 0] is 100
    max = iMax(-200, max);      assert.is(toFixed(max), 100); // max for [100, NaN, 0, -200] is 100
    max = iMax(200, max);       assert.is(toFixed(max), 200); // max for [100, NaN, -100, -200, 200] is 200
});

test('iSum()', () => {
    let sum;
    sum = iSum(10.25, undefined); assert.is(toFixed(sum), 10.25); // sum for [10.25] is 10.25
    sum = iSum(0, sum);           assert.is(toFixed(sum), 10.25); // sum for [10.25, 0] is 10.25
    sum = iSum(-0.25, sum);       assert.is(toFixed(sum), 10);    // sum for [10.25, 0, -0.25] is 10
    sum = iSum(-10, sum);         assert.is(toFixed(sum), 0);     // sum for [10.25, 0, -0.25, -10] is 0
    sum = iSum(NaN, sum);         assert.ok(isNaN(sum));          // sum for [10.25, 0, -0.25, -10, NaN] is NaN
});

test.run();