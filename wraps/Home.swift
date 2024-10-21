//
//  Home.swift
//  wraps
//
//  Created by Arjun Krishnan on 10/20/24.
//

import SwiftUI

struct CardData: Identifiable {
    static private var UNIV_ID = 0;
    var id: Int {
        CardData.UNIV_ID += 1
        return CardData.UNIV_ID
    }
    
    
}

struct HomeView: View {
    @State var cards: [CardData] = [
        CardData(),
        CardData(),
        CardData(),
    ]
    
    var body: some View {
        ScrollView {
            Text("wraps")
                .frame(maxWidth: .infinity, alignment: .leading)
                .font(.system(size: 64))
                .fontWeight(.bold)
                .padding(10)
            
            LazyVGrid(columns: [GridItem(.flexible()), GridItem(.flexible())]) {
                ForEach(self.$cards) { card in
                    VStack {}
                        .frame(maxWidth: .infinity, minHeight: 120)
                        .background(Color.red)
                        .clipShape(RoundedRectangle(cornerRadius: 10))
                }
            }
            .padding(5)
        }
    }
}
