export const flattenObject = (obj) => {
  let result = {};

  for (let key in obj) {
    if (typeof obj[key] === "object" && !Array.isArray(obj[key])) {
      const flattenedSubObject = flattenObject(obj[key]);
      for (let subKey in flattenedSubObject) {
        result[key + "." + subKey] = flattenedSubObject[subKey];
      }
    } else {
      if (key === "ready")
        result[key] = obj[key] === true ? "Ready" : "Not Ready";
      else if (key === "retired")
        result[key] = obj[key] === true ? "True" : "False";
      else if (key === "membership")
        result[key] = obj[key] === true ? "True" : "False";
      else if (key === "expiry") {
        const exp = new Date(obj[key]);
        const currentDate = new Date();
        const timeDifference = exp.getTime() - currentDate.getTime();
        const daysDifference = Math.ceil(
          timeDifference / (1000 * 60 * 60 * 24)
        );

        const options = { year: "numeric", month: "numeric", day: "numeric" };
        const expDateOnly = exp.toLocaleDateString(undefined, options);

        if (daysDifference >= 0) {
          result[key] = `${expDateOnly} (${daysDifference} days)`;
        } else {
          result[key] = `${expDateOnly} (Expired ${Math.abs(
            daysDifference
          )} days ago)`;
        }
      } else result[key] = obj[key];
    }
  }
  return result;
};

export const flattenArrayOfObject = (arr) => {
  let ans = [];
  arr.forEach((item) => {
    let obj = flattenObject(item);
    ans.push(obj);
  });
  return ans;
};
