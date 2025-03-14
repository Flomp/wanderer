import { haversineDistance } from './utils';

class GpxMetricsComputation {
  private readonly thresholdXY_m: number;  // Distance threshold for filtering on the XY axis (latitude / longitude)
  private readonly thresholdZ_m: number;  // Distance threshold for filtering on the Z axis (elevation)
  private lastFilteredPointXY: any | null = null;
  private lastFilteredZ: number | null = null;
  private lastZ: number | null = null;
  totalElevationGain = 0;
  totalElevationLoss = 0;
  totalElevationGainSmoothed = 0;
  totalElevationLossSmoothed = 0;
  totalDistance = 0;
  totalDistanceSmoothed = 0;

  constructor(thresholdXY_m: number, thresholdZ_m: number) {
    this.thresholdXY_m = thresholdXY_m;
    this.thresholdZ_m = thresholdZ_m;
  }

  addAndFilter(point: any) {
    if (!this.lastFilteredPointXY) {
      // Firstly, we are not yet filtering
      this.lastFilteredPointXY = point;
      this.lastFilteredZ = point.ele ?? 0;
      this.lastZ = point.ele ?? 0;
      return;
    }

    const distance = haversineDistance(
      this.lastFilteredPointXY.$.lat,
      this.lastFilteredPointXY.$.lon,
      point.$.lat,
      point.$.lon
    );

    this.totalDistance += distance;

    this.lastFilteredPointXY = point;

    const elevation = point.ele ?? 0;
    // @ts-ignore I know this.lastZ is not null
    const elevationDiff = elevation - this.lastZ;
    this.lastZ = elevation;
    if (elevationDiff > 0) {
      this.totalElevationGain += elevationDiff;
    }
    if (elevationDiff < 0) {
      this.totalElevationLoss -= elevationDiff;
    }

    if (distance < this.thresholdXY_m) {
      return;
    }

    this.totalDistanceSmoothed += distance;

    // @ts-ignore: I know this.lastFilteredZ is not null
    const elevationDiffSmoothed = elevation - this.lastFilteredZ;

    if (Math.abs(elevationDiffSmoothed) < this.thresholdZ_m) {
      return;
    }

    this.lastFilteredZ = elevation;
    if (elevationDiffSmoothed > 0) {
      this.totalElevationGainSmoothed += elevationDiffSmoothed;
    } else {
      this.totalElevationLossSmoothed -= elevationDiffSmoothed;
    }
  }
}

export default GpxMetricsComputation;
