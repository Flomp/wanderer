export { getFitMessage, getFitMessageBaseType }

import { getFieldObject, getMessageName } from "./fit"

function getFitMessage(messageNum: any) {
  return {
    name: getMessageName(messageNum),
    getAttributes: function getAttributes(fieldNum: any) {
      return getFieldObject(fieldNum, messageNum);
    }
  };
}

function getFitMessageBaseType(foo: any) {
  return foo;
}