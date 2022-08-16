import { ClrDatagridComparatorInterface } from '@clr/angular';

export class CustomComparator implements ClrDatagridComparatorInterface<any> {

    fieldName: string;
    type: string;

    constructor(fieldName: string, type: string) {
        this.fieldName = fieldName;
        this.type = type;
    }

    compare(a: { [key: string]: any | any[] }, b: { [key: string]: any | any[] }) {
        let comp = 0;
        if (a && b) {
            let fieldA, fieldB;
            for (let key of Object.keys(a)) {
                if (key === this.fieldName) {
                    fieldA = a[key];
                    fieldB = b[key];
                    break;
                } else if (typeof a[key] === 'object') {
                    let insideObject = a[key];
                    for (let insideKey in insideObject) {
                        if (insideKey === this.fieldName) {
                            fieldA = insideObject[insideKey];
                            fieldB = b[key][insideKey];
                            break;
                        }
                    }
                }
            }
            switch (this.type) {
                case "number":
                    comp = fieldB - fieldA;
                    break;
                case "date":
                    comp = new Date(fieldB).getTime() - new Date(fieldA).getTime();
                    break;
                case "string":
                    comp = fieldB.localeCompare(fieldA);
                    break;
                //project job management list defualt sorting
                //firstly put "pending on this site" job on top, then sort others by creation time, if there are multiple "pending on this site" jobs, they should be sorted by creation time too
                case "job":
                    if (a.pending_on_this_site && !b.pending_on_this_site) {
                        comp = -1
                    } else if (!a.pending_on_this_site && b.pending_on_this_site) {
                        comp = 1
                    } else {
                        comp = new Date(fieldB).getTime() - new Date(fieldA).getTime();
                    }
                    break;
            }
        }
        return comp;
    }
}