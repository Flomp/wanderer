import { parse } from "date-fns";
import { date, number, object, string } from "yup";

class SummitLog {
  id?: string;
  date: string;
  text?: string;

  constructor(date: string, params?: { id?: string, text?: string }) {
    this.date = date;
    this.id = params?.id;
    this.text = params?.text ?? "";
  }
}

const summitLogSchema = object<SummitLog>({
  id: string().optional(),
  date: date().transform((value, originalValue, context) => {
    if (context.isType(value)) return value;
    return parse(originalValue, 'dd.MM.yyyy', new Date());
  }).required('Required').typeError('Invalid Date'),
  text: string().optional(),
});


export { SummitLog, summitLogSchema }