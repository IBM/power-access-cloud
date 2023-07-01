const clientSearchFilter = (searchText, data) => {
  if (!searchText) return data;
  let cleanedSearchText = searchText;
  const specialChars = "!@#$^&%*()+=[]/\\{}|:<>?,";

  for (var i = 0; i < specialChars.length; i++) {
    cleanedSearchText = cleanedSearchText.replace(
      new RegExp("\\" + specialChars[i], "gi"),
      ""
    );
  }

  const re = new RegExp(cleanedSearchText, "gim");
  const filteredRows = data.filter((x) => {
    const keys = Object.keys(x).filter((key) => {
      const value = x[key];
      if (!value) return false;
      switch (typeof value) {
        case "string":
          return value.match(re);
        case "boolean":
          return Boolean.toString(value).match(re);
        case "number":
          return `${value}`.match(re);
        case "object":
          if (Array.isArray(value)) {
            const isMatch = value.some((x) => re.test(x));
            return isMatch;
          } else {
            return false;
          }
        default:
          return false;
      }
    });
    if (keys.length > 0) {
      return x;
    } else {
      return null;
    }
  });

  return filteredRows;
};

export { clientSearchFilter };
