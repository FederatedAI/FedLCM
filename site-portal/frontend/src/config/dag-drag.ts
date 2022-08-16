import * as dagreD3 from 'dagre-d3'
import * as d3 from 'd3'
export default class Dag {
    g: any;
    dagModify: any
    that: any
    node_shape = "rect";
    node_stype = "fill:#fff;stroke:#000;";
    edge_stype = "fill:#fff;stroke:#333;stroke-width:1.5px";
    tooltip_css = "tooltip";
    rankdir = "TB";
    attrForm: { [key: string]: any } = {
        moduleName: '',
        options: 'common',
        attrList: [],
        diffList: [],
        flag: true
    }
    startStr = 'reader_0'
    dsl: any;
    canva_selector: any;
    conf: any;
    d3: any
    zoomMultiples = 1
    constructor(dsl: any, canva_selector: any, callback: Function, that: any) {
        // data source
        this.that = that
        this.dsl = dsl;
        this.dagModify = callback
        for (const key in dsl) {
            if (key === this.startStr) {
                this.attrForm.moduleName = this.startStr
                for (const attr in dsl[key].attributes) {
                    if (this.dsl[key].attributes[attr].hasOwnProperty('drop_down_box')) {
                                
                        this.attrForm.attrList.push({
                            name: attr,
                            value: this.dsl[key].attributes[attr]
                        })
                    } else {
                        this.attrForm.attrList.push({
                            name: attr,
                            value: JSON.stringify(this.dsl[key].attributes[attr])
                        })
                    }
                    // this.attrForm.attrList.push({
                    //     name: attr,
                    //     value: dsl[key][attr],
                    // })
                }
            }
        }
        this.canva_selector = canva_selector
        this.d3 = d3
    }

    get_source_data(input_filed: any): any {
        var source_inputs = new Array();
        for (var key in input_filed) {
            if (!Array.isArray(input_filed[key])) {
                return this.get_source_data(input_filed[key]);
            } else {
                input_filed[key].forEach((val: any) => {
                    source_inputs.push(val);
                });
            }
        }
        return source_inputs
    }

    handle_common_components(component_name: any, common_components: any, results: any) {
        var iter = function (obj: any, results: any) {
            for (var key in obj) {
                if (key == component_name) {
                    results.push(obj[key])
                    return results;
                } else {
                    if (typeof obj[key] == "object") {
                        iter(obj[key], results);
                    } else {
                        return results;
                    }
                }
            }

            return results
        }

        return iter(common_components, results)
    }

    handle_role_components(component_name: any, role_components: any, results: any) {
        var found = false
        var iter = function (obj: any, results: any) {
            for (var key in obj) {
                if (key == component_name) {
                    found = true
                    results.push(obj[key])
                    return results;
                } else {
                    if (key == "guest" || key == "host") {
                        results.push({
                            "role": key
                        })
                    }

                    if (typeof obj[key] == "object") {
                        iter(obj[key], results);
                    } else {
                        return results;
                    }
                }
            }
            if (!found) {
                results.pop("role")
            }

            return results
        }

        return iter(role_components, results)
    }

    get_component_para(component_name: any) {
        var components_paras = this.conf["component_parameters"];
        var results = new Array();
        if (components_paras.hasOwnProperty("common")) {
            this.handle_common_components(component_name, components_paras["common"], results)
        }

        if (results.length == 0 && components_paras.hasOwnProperty("role")) {
            this.handle_role_components(component_name, components_paras["role"], results)
        }

        return results
    }

    print_parameter(parameters: any) {
        var parameters_str = ""
        parameters.forEach((onepara: any) => {
            var str = "<dl>"
            for (var key in onepara) {
                if (key == "role") {
                    str += "Role: " + onepara[key]
                } else {
                    str += "<dt>" + key + "</dt>"
                    str += "<dd>" + JSON.stringify(onepara[key]) + "</dd>"
                }
            }
            str += "</dl>"
            parameters_str += str
        });
        return parameters_str;
    }

    Generate() {
        this.g = new dagreD3.graphlib.Graph();
        this.g.setGraph({
            rankdir: this.rankdir
        });

        for (var component_name in this.dsl) {
            // get algorithm
            var one_model = this.dsl[component_name]
            this.g.setNode(component_name, {
                label: component_name,
                rx: 5,
                ry: 5,
                width: 130,
                shape: this.node_shape,
                style: this.node_stype,
                x: '50%',
                y: '50%'
            });

        }
        for (var component_name in this.dsl) {
            var one_model = this.dsl[component_name];
            // decide if the node is root
            if (one_model.conditions.hasOwnProperty("input")) {
                var inputs = this.get_source_data(one_model.conditions["input"])
                inputs.forEach((one_input: any) => {
                    var source_component = one_input.split(".")[0];
                    const result = Object.keys(this.dsl).find(el => el === source_component)
                    if (result) {
                        this.getComponentRelation(component_name, inputs[0].split(".")[0])
                        this.g.setEdge(source_component, component_name, {
                            label: "",
                            style: this.edge_stype
                        });
                    }
                });
            } else {
                this.checkboxFirst = component_name
            }
        }
        this.my[0] = {
            name: 'reader_0',
            value: false
        }
        this.level = 0
        this.count = 0
        this.mapCheckbox(this.checkboxFirst)

    }
    checkboxFirst = ''
    my: any[] = []
    level = 0
    count = 0
    Draw() {
        // clear board
        d3.selectAll("#svg-canvas-drop > *").remove();

        let render = new dagreD3.render();
        // getsvg
        let svg = this.d3.select(this.canva_selector);
        let svgGroup: any = svg.append('g').attr('transform', 'translate(310,0)scale(' + this.zoomMultiples + ')')
        render(svgGroup, this.g);
        let tooltip = this.d3.select("body").append("div").classed(this.tooltip_css, true).style("opacity", 0).style("display", "none");
        svg.selectAll("g.node").on('click', (name: string) => {
            this.attrForm.moduleName = ''
            this.attrForm.attrList = []
            this.attrForm.diffList = []
            for (const key in this.dsl) {
                if (key === name) {
                    this.attrForm.moduleName = name
                    for (const attr in this.dsl[key].attributes) {
                        if (Object.prototype.toString.call(this.dsl[key].attributes[attr]) === '[object Object]'
                            || Object.prototype.toString.call(this.dsl[key].attributes[attr]) === '[object Array]') {
                            if (this.dsl[key].attributes[attr].hasOwnProperty('drop_down_box')) {

                                this.attrForm.attrList.push({
                                    name: attr,
                                    value: this.dsl[key].attributes[attr]
                                })
                            } else {
                                this.attrForm.attrList.push({
                                    name: attr,
                                    value: JSON.stringify(this.dsl[key].attributes[attr])
                                })
                            }
                        } else {
                            this.attrForm.attrList.push({
                                name: attr,
                                value: this.dsl[key].attributes[attr],
                            })
                        }
                    }
                    if (this.dsl[key].attributeType === 'common') {
                        this.attrForm.options = 'common'
                    } else {
                        this.attrForm.options = 'diff'
                        for (const attr in this.dsl[key].diffAttribute) {
                            this.attrForm.diffList.push({
                                name: attr,
                                form: this.dsl[key].diffAttribute[attr]
                            })
                        }
                    }
                }
            }
            this.dagModify(this.attrForm, this.that)
        })
        svg.selectAll("g.node").on('dblclick', (name: string) => {
            this.that.dragObj = this.dsl[name]
            this.that.bulletFrame(false, name)
        })
        svg.selectAll("g.node").append('image').attr('xlink:href', '../assets/close.jpg').attr('width', 10).attr('fill', 'none').attr('transform', 'translate(75,-17)')
        svg.selectAll("g.node image").on("click", (e: any, a: any, c: any) => {
            this.g.removeNode(e)
            this.g.removeEdge(e)
            delete this.dsl[e]
            this.cObj = {}
            this.my.forEach((el, index) => {
                if (el.name === e) {
                    this.my.splice(index, 1)
                    this.that.program.splice(index, 1)
                }
            })
            render(svgGroup, this.g);
        })
        svg.selectAll("g.edgePath").on("mouseover", (e: any) => {
        });
        svg.attr("preserveAspectRatio", "xMidYMin")
        if (this.count > 3 && this.level < 11) {
            svg.attr("viewBox", "0 0 " + ((this.count - 3) * 200 + 900) + " 783")
        } else if (this.count > 3 && this.level > 10) {
            svg.attr("viewBox", "0 0 " + ((this.count - 3) * 200 + 900) + " " + ((this.level - 10) * 90 + 783))
        } else if (this.count < 3 && this.level > 10) {
            svg.attr("viewBox", "0 0 900 " + ((this.level - 10) * 90 + 783))
        } else {
            svg.attr("viewBox", "0 0 900 783")
        }
    }
    cObj: any = {}
    getComponentRelation(child: string, component_name: string) {
        if (this.cObj.hasOwnProperty(component_name)) {
            if (!this.cObj[component_name].children.find((el: string) => el === child)) {
                this.cObj[component_name].children.push(child)
            }
            if (!this.cObj.hasOwnProperty(child)) {
                this.cObj[child] = {
                    name: child,
                    children: []
                }
            }
        } else {
            this.cObj[component_name] = {
                name: component_name,
                children: [child]
            }
        }
    }
    mapCheckbox(name: string): any {
        if (!this.cObj[name] || !(this.cObj[name].children.length > 0)) {
            this.level++
            if (this.find(name)) {
                this.my.push({
                    name: name,
                    value: false,
                })
            }
            this.checkboxFirst = name
            return false
        } else if (this.cObj[name].children.length < 2) {
            this.level++
            if (this.find(this.cObj[name].children[0])) {
                this.my.push({
                    name: this.cObj[name].children[0],
                    value: false
                })
            }
            this.checkboxFirst = this.cObj[name].children[0]
            return this.mapCheckbox(this.checkboxFirst)
        } else if (this.cObj[name].children.length > 2) {
            this.level++
            if (this.cObj[name].children.length > this.count) this.count = this.cObj[name].children.length
            this.cObj[name].children.forEach((el: string) => {
                if (this.find(el)) {
                    this.my.push({
                        name: el,
                        value: false
                    })
                }
            });
            this.cObj[name].children.forEach((el: string) => {
                this.checkboxFirst = el
                return this.mapCheckbox(this.checkboxFirst)
            });
        }
    }
    find(item: any) {
        return this.my.findIndex(c => c.name === item) === -1 ? true : false
    }
}