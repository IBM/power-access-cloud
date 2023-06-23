export const flattenObject = (obj)=> {
    let result = {};
  
    for (let key in obj) {
        if (typeof obj[key] === 'object' && !Array.isArray(obj[key])) {
            const flattenedSubObject = flattenObject(obj[key]);
            for (let subKey in flattenedSubObject) {
            result[key + '.' + subKey] = flattenedSubObject[subKey];
            }
        } else {
            if (key === "ready")
                result[key] = ((obj[key] === true) ? "Ready" : "Not Ready");
            else
            result[key] = obj[key];
        }
    }   
    return result;
}

export const flattenArrayOfObject = (arr)=>{
    let ans = [];
    arr.forEach((item)=>{
        let obj = flattenObject(item);
        ans.push(obj);
    })
    return ans;
};