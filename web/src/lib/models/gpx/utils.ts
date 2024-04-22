function removeEmpty(obj: Record<string, any>) {
  Object.entries(obj).forEach(([key, val]) => {
    if (val && val instanceof Object) {
      removeEmpty(val);
    } else if (val == null) {
      delete obj[key];
    }
  });
}

function allDatesToISOString(obj: Record<string, any>) {
  Object.entries(obj).forEach(([key, val]) => {
    if (val) {
      if (val instanceof Date) {
        obj[key] = val.toISOString().split('.')[0] + 'Z';
      } else if (val instanceof Object) {
        allDatesToISOString(val);
      }
    }
  });
}

export { removeEmpty, allDatesToISOString };
